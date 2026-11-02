package repositories

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestDocumentosDisco(t *testing.T) {
	store, err := NewDocumentoStoreDisco(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pdf := []byte("%PDF-1.4 ... %%EOF")

	id, err := store.Guardar(ctx, 7, `C:\Users\maria\Título de Maestría.pdf`, pdf)
	if err != nil {
		t.Fatal(err)
	}
	if !IDDocumentoValido(id) || !strings.HasSuffix(id, "__Titulo-de-Maestria.pdf") {
		t.Errorf("id inesperado: %q", id)
	}

	datos, err := store.Leer(ctx, 7, id)
	if err != nil || !bytes.Equal(datos, pdf) {
		t.Errorf("el documento leído no coincide (err=%v)", err)
	}
	if existe, _ := store.Existe(ctx, 7, id); !existe {
		t.Error("el documento recién guardado debe existir")
	}

	// Otro postulante NO puede leer el documento aunque conozca el id.
	if _, err := store.Leer(ctx, 8, id); !errors.Is(err, ErrNoEncontrado) {
		t.Error("un postulante no debe poder leer documentos de otro")
	}
	// Rutas maliciosas se rechazan sin tocar el disco.
	for _, malo := range []string{"../../.env", "..\\..\\main.go", "x.pdf", id + "/../otro.pdf"} {
		if existe, _ := store.Existe(ctx, 7, malo); existe {
			t.Errorf("el id %q no debería aceptarse", malo)
		}
	}
}
