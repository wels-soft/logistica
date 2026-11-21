package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type FieldMeta struct {
	Name  string
	Label string
	Type  string
	Value string
}

type EntityConfig struct {
	Slug      string
	Title     string
	ModelType reflect.Type
}

var db *gorm.DB

var entities = []EntityConfig{
	{Slug: "sucursales", Title: "Sucursales", ModelType: reflect.TypeOf(Sucursal{})},
	{Slug: "vehiculos", Title: "Vehículos", ModelType: reflect.TypeOf(Vehiculo{})},
	{Slug: "conductores", Title: "Conductores", ModelType: reflect.TypeOf(Conductor{})},
	{Slug: "clientes", Title: "Clientes", ModelType: reflect.TypeOf(Cliente{})},
	{Slug: "envios", Title: "Envíos", ModelType: reflect.TypeOf(Envio{})},
}

func main() {
	dsn := "host=10.0.2.47 user=logistica_user password=Logistica2026 dbname=logistica_db port=5432 sslmode=disable TimeZone=America/Guayaquil"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error conectando a PostgreSQL: %v", err)
	}

	// Scaffolding de Base de Datos: crea las 5 tablas y sus relaciones FK
	err = db.AutoMigrate(&Sucursal{}, &Vehiculo{}, &Conductor{}, &Cliente{}, &Envio{})
	if err != nil {
		log.Fatalf("Error en AutoMigrate: %v", err)
	}

	// Crear llaves foráneas explícitas en PostgreSQL
	db.Exec("ALTER TABLE vehiculos ADD CONSTRAINT fk_vehiculos_sucursal FOREIGN KEY (sucursal_id) REFERENCES sucursals(id) ON DELETE RESTRICT;")
	db.Exec("ALTER TABLE conductors ADD CONSTRAINT fk_conductores_sucursal FOREIGN KEY (sucursal_id) REFERENCES sucursals(id) ON DELETE RESTRICT;")
	db.Exec("ALTER TABLE envios ADD CONSTRAINT fk_envios_vehiculo FOREIGN KEY (vehiculo_id) REFERENCES vehiculos(id) ON DELETE RESTRICT;")
	db.Exec("ALTER TABLE envios ADD CONSTRAINT fk_envios_conductor FOREIGN KEY (conductor_id) REFERENCES conductors(id) ON DELETE RESTRICT;")
	db.Exec("ALTER TABLE envios ADD CONSTRAINT fk_envios_cliente FOREIGN KEY (cliente_id) REFERENCES clientes(id) ON DELETE RESTRICT;")

	e := echo.New()
	e.Renderer = &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	// Ruta principal (Dashboard con resumen de las 5 entidades)
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/sucursales")
	})

	// Scaffolding automático de rutas CRUD para cada una de las 5 entidades
	for _, ent := range entities {
		registerScaffoldRoutes(e, ent)
	}

	log.Println("Servidor Logística corriendo en http://localhost:8085")
	e.Logger.Fatal(e.Start(":8085"))
}

func registerScaffoldRoutes(e *echo.Echo, cfg EntityConfig) {
	base := "/" + cfg.Slug

	// LISTAR (Index)
	e.GET(base, func(c echo.Context) error {
		slicePtr := reflect.New(reflect.SliceOf(cfg.ModelType))
		db.Order("id asc").Find(slicePtr.Interface())

		headers := getHeaders(cfg.ModelType)
		rows := extractRows(slicePtr.Elem())

		return c.Render(http.StatusOK, "crud.html", map[string]interface{}{
			"Entities":   entities,
			"Current":    cfg,
			"Headers":    headers,
			"Rows":       rows,
			"FormFields": buildFormFields(cfg.ModelType, nil),
			"EditID":     "",
			"Error":      c.QueryParam("error"),
		})
	})

	// CREAR (Create)
	e.POST(base, func(c echo.Context) error {
		itemPtr := reflect.New(cfg.ModelType)
		bindFormToStruct(c, itemPtr.Elem())
		if res := db.Create(itemPtr.Interface()); res.Error != nil {
			return c.Redirect(http.StatusFound, base+"?error="+res.Error.Error())
		}
		return c.Redirect(http.StatusFound, base)
	})

	// VISTA EDITAR (Edit)
	e.GET(base+"/edit/:id", func(c echo.Context) error {
		id := c.Param("id")
		itemPtr := reflect.New(cfg.ModelType)
		if err := db.First(itemPtr.Interface(), id).Error; err != nil {
			return c.Redirect(http.StatusFound, base)
		}

		slicePtr := reflect.New(reflect.SliceOf(cfg.ModelType))
		db.Order("id asc").Find(slicePtr.Interface())

		return c.Render(http.StatusOK, "crud.html", map[string]interface{}{
			"Entities":   entities,
			"Current":    cfg,
			"Headers":    getHeaders(cfg.ModelType),
			"Rows":       extractRows(slicePtr.Elem()),
			"FormFields": buildFormFields(cfg.ModelType, itemPtr.Elem().Interface()),
			"EditID":     id,
			"Error":      "",
		})
	})

	// ACTUALIZAR (Update)
	e.POST(base+"/update/:id", func(c echo.Context) error {
		id := c.Param("id")
		itemPtr := reflect.New(cfg.ModelType)
		if err := db.First(itemPtr.Interface(), id).Error; err == nil {
			bindFormToStruct(c, itemPtr.Elem())
			if res := db.Save(itemPtr.Interface()); res.Error != nil {
				return c.Redirect(http.StatusFound, base+"?error="+res.Error.Error())
			}
		}
		return c.Redirect(http.StatusFound, base)
	})

	// ELIMINAR (Delete)
	e.POST(base+"/delete/:id", func(c echo.Context) error {
		id := c.Param("id")
		itemPtr := reflect.New(cfg.ModelType)
		if res := db.Delete(itemPtr.Interface(), id); res.Error != nil {
			return c.Redirect(http.StatusFound, base+"?error=No se puede eliminar: tiene registros relacionados")
		}
		return c.Redirect(http.StatusFound, base)
	})
}

func getHeaders(t reflect.Type) []string {
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		headers = append(headers, t.Field(i).Tag.Get("label"))
	}
	return headers
}

func extractRows(slice reflect.Value) [][]string {
	var rows [][]string
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i)
		var row []string
		for j := 0; j < item.NumField(); j++ {
			row = append(row, fmt.Sprintf("%v", item.Field(j).Interface()))
		}
		rows = append(rows, row)
	}
	return rows
}

func buildFormFields(t reflect.Type, existing interface{}) []FieldMeta {
	var fields []FieldMeta
	var val reflect.Value
	if existing != nil {
		val = reflect.ValueOf(existing)
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == "ID" {
			continue // ID es autoincremental
		}
		inputType := "text"
		if f.Type.Kind() == reflect.Int || f.Type.Kind() == reflect.Uint {
			inputType = "number"
		}
		currentVal := ""
		if existing != nil {
			currentVal = fmt.Sprintf("%v", val.Field(i).Interface())
		}
		fields = append(fields, FieldMeta{
			Name:  f.Name,
			Label: f.Tag.Get("label"),
			Type:  inputType,
			Value: currentVal,
		})
	}
	return fields
}

func bindFormToStruct(c echo.Context, structVal reflect.Value) {
	t := structVal.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == "ID" {
			continue
		}
		raw := c.FormValue(f.Name)
		field := structVal.Field(i)
		switch field.Kind() {
		case reflect.String:
			field.SetString(raw)
		case reflect.Int:
			v, _ := strconv.Atoi(raw)
			field.SetInt(int64(v))
		case reflect.Uint:
			v, _ := strconv.ParseUint(raw, 10, 64)
			field.SetUint(v)
		}
	}
}
