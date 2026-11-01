package services

import "testing"

func TestCedulaValida(t *testing.T) {
	casos := []struct {
		cedula   string
		esperado bool
	}{
		{"0901234567", true},  // datos semilla
		{"1702345677", true},
		{"2403456789", true},
		{"0104567896", true},
		{"1705678900", true},
		{"0901234568", false}, // dígito verificador incorrecto
		{"090123456", false},  // 9 dígitos
		{"09012345678", false},
		{"abcdefghij", false},
		{"9901234567", false}, // provincia inexistente
		{"0961234567", false}, // tercer dígito > 5
		{"", false},
	}
	for _, c := range casos {
		if got := CedulaValida(c.cedula); got != c.esperado {
			t.Errorf("CedulaValida(%q) = %v, se esperaba %v", c.cedula, got, c.esperado)
		}
	}
}

func TestParseFecha(t *testing.T) {
	for _, s := range []string{"21/06/2018", "2018-06-21", " 21/06/2018 "} {
		if _, err := parseFecha(s); err != nil {
			t.Errorf("parseFecha(%q) no debería fallar: %v", s, err)
		}
	}
	for _, s := range []string{"", "2018/06/21", "31/02/2018", "hola"} {
		if _, err := parseFecha(s); err == nil {
			t.Errorf("parseFecha(%q) debería fallar", s)
		}
	}
}

func TestTextoTiempo(t *testing.T) {
	casos := map[int]string{600: "10 minutos", 601: "11 minutos", 60: "1 minuto", 5: "1 minuto", 120: "2 minutos"}
	for seg, esperado := range casos {
		if got := textoTiempo(seg); got != esperado {
			t.Errorf("textoTiempo(%d) = %q, se esperaba %q", seg, got, esperado)
		}
	}
}

func TestValidarPassword(t *testing.T) {
	if validarPassword("Clave1234") != nil {
		t.Error("Clave1234 debería ser válida")
	}
	for _, p := range []string{"corta1", "sololetrasss", "123456789", ""} {
		if validarPassword(p) == nil {
			t.Errorf("%q debería ser rechazada", p)
		}
	}
}
