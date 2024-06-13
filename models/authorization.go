package models

type Role string

const (
	RoleBinusian Role = "BINUSIAN"
	RoleUMKM     Role = "UMKM"
)

type Level string

const (
	LevelAdmin Level = "admin"
	LevelUser  Level = "user"
)
