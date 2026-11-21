package main

// 1. Entidad Sucursal (5 atributos)
type Sucursal struct {
	ID        uint   `gorm:"primaryKey" label:"ID"`
	Nombre    string `gorm:"size:100;not null" label:"Nombre de Sucursal"`
	Ciudad    string `gorm:"size:80;not null" label:"Ciudad"`
	Direccion string `gorm:"size:150" label:"Dirección"`
	Telefono  string `gorm:"size:20" label:"Teléfono"`
}

// 2. Entidad Vehiculo (5 atributos - Relacionada con Sucursal)
type Vehiculo struct {
	ID          uint   `gorm:"primaryKey" label:"ID"`
	Placa       string `gorm:"size:20;unique;not null" label:"Placa"`
	Marca       string `gorm:"size:60;not null" label:"Marca"`
	CapacidadKg int    `gorm:"not null" label:"Capacidad (Kg)"`
	SucursalID  uint   `gorm:"not null" label:"ID Sucursal (FK)"`
}

// 3. Entidad Conductor (5 atributos - Relacionada con Sucursal)
type Conductor struct {
	ID         uint   `gorm:"primaryKey" label:"ID"`
	Nombre     string `gorm:"size:100;not null" label:"Nombre Completo"`
	Licencia   string `gorm:"size:30;unique;not null" label:"Nro. Licencia"`
	Telefono   string `gorm:"size:20" label:"Teléfono"`
	SucursalID uint   `gorm:"not null" label:"ID Sucursal (FK)"`
}

// 4. Entidad Cliente (5 atributos)
type Cliente struct {
	ID            uint   `gorm:"primaryKey" label:"ID"`
	RazonSocial   string `gorm:"size:120;not null" label:"Razón Social"`
	RUC           string `gorm:"size:20;unique;not null" label:"RUC / Cédula"`
	Email         string `gorm:"size:100" label:"Correo Electrónico"`
	CiudadDestino string `gorm:"size:80;not null" label:"Ciudad Destino"`
}

// 5. Entidad Envio (5 atributos - Relaciona Vehiculo, Conductor y Cliente)
type Envio struct {
	ID          uint   `gorm:"primaryKey" label:"ID"`
	CodigoGuia  string `gorm:"size:30;unique;not null" label:"Código de Guía"`
	VehiculoID  uint   `gorm:"not null" label:"ID Vehículo (FK)"`
	ConductorID uint   `gorm:"not null" label:"ID Conductor (FK)"`
	ClienteID   uint   `gorm:"not null" label:"ID Cliente (FK)"`
}
