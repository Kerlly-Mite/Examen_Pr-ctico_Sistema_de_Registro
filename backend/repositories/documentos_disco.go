package repositories

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// documentosDisco guarda cada PDF en <dir>/<postulante_id>/<id>. La carpeta del postulante es lo que
// garantiza que nadie pueda leer los documentos de otra persona.
type documentosDisco struct {
	dir string
}

// NewDocumentoStoreDisco crea (si no existe) la carpeta base de los documentos.
func NewDocumentoStoreDisco(dir string) (DocumentoStore, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("crear carpeta de documentos %s: %w", dir, err)
	}
	return &documentosDisco{dir: dir}, nil
}

// El id tiene 32 caracteres aleatorios + "__" + el nombre original limpio, p. ej.
// "9f2c...e1__titulo-maestria.pdf". Así el usuario reconoce su archivo y el id no se puede adivinar.
var (
	reIDDocumento   = regexp.MustCompile(`^[a-f0-9]{32}__[A-Za-z0-9._-]{1,120}\.pdf$`)
	reCaracterLibre = regexp.MustCompile(`[^A-Za-z0-9._-]+`)
)

// IDDocumentoValido evita rutas maliciosas como "../../.env" (path traversal).
func IDDocumentoValido(id string) bool { return reIDDocumento.MatchString(id) }

func (d *documentosDisco) Guardar(_ context.Context, postulanteID int64, nombreOriginal string, datos []byte) (string, error) {
	var r [16]byte
	if _, err := rand.Read(r[:]); err != nil {
		return "", fmt.Errorf("generar id de documento: %w", err)
	}
	id := hex.EncodeToString(r[:]) + "__" + nombreLimpio(nombreOriginal) + ".pdf"

	carpeta := d.carpeta(postulanteID)
	if err := os.MkdirAll(carpeta, 0o750); err != nil {
		return "", fmt.Errorf("crear carpeta del postulante: %w", err)
	}
	// Se escribe en un archivo temporal y luego se renombra: nunca queda un PDF a medio escribir.
	tmp, err := os.CreateTemp(carpeta, "subida-*.tmp")
	if err != nil {
		return "", fmt.Errorf("crear archivo temporal: %w", err)
	}
	if _, err := tmp.Write(datos); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("escribir documento: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("cerrar documento: %w", err)
	}
	if err := os.Rename(tmp.Name(), filepath.Join(carpeta, id)); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("guardar documento: %w", err)
	}
	return id, nil
}

func (d *documentosDisco) Leer(_ context.Context, postulanteID int64, id string) ([]byte, error) {
	if !IDDocumentoValido(id) {
		return nil, ErrNoEncontrado
	}
	datos, err := os.ReadFile(filepath.Join(d.carpeta(postulanteID), id))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("leer documento: %w", err)
	}
	return datos, nil
}

func (d *documentosDisco) Existe(_ context.Context, postulanteID int64, id string) (bool, error) {
	if !IDDocumentoValido(id) {
		return false, nil
	}
	info, err := os.Stat(filepath.Join(d.carpeta(postulanteID), id))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("verificar documento: %w", err)
	}
	return info.Mode().IsRegular(), nil
}

func (d *documentosDisco) Listar(_ context.Context, postulanteID int64) ([]DocumentoInfo, error) {
	entradas, err := os.ReadDir(d.carpeta(postulanteID))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil // aún no ha subido nada
	}
	if err != nil {
		return nil, fmt.Errorf("listar documentos: %w", err)
	}
	var docs []DocumentoInfo
	for _, e := range entradas {
		if !e.Type().IsRegular() || !IDDocumentoValido(e.Name()) { // ignora temporales a medio subir
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, fmt.Errorf("leer datos del documento: %w", err)
		}
		docs = append(docs, DocumentoInfo{ID: e.Name(), Tamano: info.Size(), Modificado: info.ModTime()})
	}
	return docs, nil
}

func (d *documentosDisco) carpeta(postulanteID int64) string {
	return filepath.Join(d.dir, strconv.FormatInt(postulanteID, 10))
}

var sinTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ü", "U", "Ñ", "N", " ", "-")

// nombreLimpio deja solo letras sin tilde, números, punto, guion y guion bajo, sin la extensión .pdf.
func nombreLimpio(nombre string) string {
	nombre = sinTildes.Replace(filepath.Base(strings.ReplaceAll(nombre, "\\", "/")))
	nombre = strings.TrimSuffix(strings.TrimSuffix(nombre, ".pdf"), ".PDF")
	nombre = strings.Trim(reCaracterLibre.ReplaceAllString(nombre, "-"), "-.")
	if nombre == "" {
		nombre = "documento"
	}
	if len(nombre) > 80 {
		nombre = nombre[:80]
	}
	return nombre
}
