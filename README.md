# Ficha Tracker

Aplicativo desktop para registrar a impressão de fichas de saúde (nome, tipo de
solicitação e ACS), com data e hora, para saber rapidamente se uma ficha já foi
impressa.

## Stack

- **Backend**: Go, em camadas (`domain` → `service` → `repository` → handler `Wails`)
- **Persistência**: SQLite local (arquivo em `~/.config/ficha-tracker/fichas.db` no Linux)
- **Frontend**: React (Vite)
- **Empacotamento**: Wails (executável único, sem instalação)

## Funcionalidades

- Registrar ficha (nome, tipo de solicitação, ACS)
- Listar todas as fichas registradas (mais recentes primeiro)
- Buscar fichas por nome

## Rodando em desenvolvimento

Em sistemas com apenas `libwebkit2gtk-4.1` (Ubuntu 24.04+/25+), é preciso instalar o
pacote de desenvolvimento e passar a tag de build correspondente:

```bash
sudo apt install libwebkit2gtk-4.1-dev
wails dev -tags webkit2_41
```

Em sistemas com `libwebkit2gtk-4.0-dev` disponível, basta `wails dev`.

## Build de produção

```bash
wails build -tags webkit2_41
```

Gera o executável em `build/bin/ficha-tracker`. Para gerar o `.exe` do Windows,
compile em uma máquina Windows (ou via CI com o target apropriado) — o backend usa
`modernc.org/sqlite` (driver puro Go, sem CGO), então não há dependências nativas
extras para essa parte.

## Testes

```bash
go test ./...
```

## Arquitetura

```
React UI
  ↓ IPC
App (handler Wails, internal em app.go)
  ↓
FichaService (internal/service) — validações e regras
  ↓
FichaRepository (internal/repository) — SQLite
  ↓
Ficha (internal/domain) — modelo puro + validação
```
