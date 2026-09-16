package domain

// RequestTypes is the canonical list of ficha request types offered by the
// UI. "Outro" is a catch-all: when chosen, the caller supplies free text as
// the actual RequestType value on the ficha instead of "Outro" itself.
var RequestTypes = []string{
	"Exame de Sangue",
	"UCG",
	"RX",
	"Pediatra",
	"Neurologista",
	"Fisioterapeuta",
	"Oftalmologista",
	"Mastologista",
	"Psicólogo",
	"Cardiologista",
	"Dermatologista",
	"Urologista",
	"Otorrinolaringologista",
	"Gastroenterologista",
	"Nutricionista",
	"Angiologista",
	"Ginecologista",
	"Endocrinologista",
	"Reumatologista",
	"Ortopedista",
	"Psiquiatra",
	"Pneumologista",
	"Obstetra",
	"Outro",
}
