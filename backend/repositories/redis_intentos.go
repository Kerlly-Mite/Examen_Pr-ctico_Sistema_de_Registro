package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Script Lua: se ejecuta de forma ATÓMICA en Redis (dos peticiones simultáneas no pueden saltarse el contador).
//   KEYS[1] = contador de intentos      KEYS[2] = marca de bloqueo
//   ARGV[1] = máximo de intentos        ARGV[2] = TTL del contador (s)     ARGV[3] = duración del bloqueo (s)
// Devuelve {intentos, bloqueado(0/1)}.
var scriptFallo = redis.NewScript(`
local n = redis.call('INCR', KEYS[1])
if n == 1 then redis.call('EXPIRE', KEYS[1], ARGV[2]) end
if n >= tonumber(ARGV[1]) then
  redis.call('SET', KEYS[2], '1', 'EX', ARGV[3])
  redis.call('DEL', KEYS[1])
  return {n, 1}
end
return {n, 0}
`)

type limitadorRedis struct {
	client      *redis.Client
	maxIntentos int
	bloqueo     time.Duration
}

// NewLimitadorIntentos crea el limitador. Con maxIntentos=3 y bloqueo=10min cumple EF-11:
// al tercer intento fallido consecutivo se bloquea 10 minutos; pasado ese tiempo Redis borra la llave sola (TTL).
func NewLimitadorIntentos(client *redis.Client, maxIntentos int, bloqueo time.Duration) LimitadorIntentos {
	return &limitadorRedis{client: client, maxIntentos: maxIntentos, bloqueo: bloqueo}
}

func claveIntentos(contexto, id string) string { return fmt.Sprintf("intentos:%s:%s", contexto, id) }
func claveBloqueo(contexto, id string) string  { return fmt.Sprintf("bloqueo:%s:%s", contexto, id) }

func (l *limitadorRedis) Verificar(ctx context.Context, contexto, id string) (int, error) {
	ttl, err := l.client.TTL(ctx, claveBloqueo(contexto, id)).Result()
	if err != nil {
		return 0, fmt.Errorf("consultar bloqueo: %w", err)
	}
	if ttl > 0 {
		// Se redondea hacia arriba para no mostrar "0 segundos" mientras sigue bloqueado.
		return int((ttl + time.Second - 1) / time.Second), nil
	}
	return 0, nil
}

func (l *limitadorRedis) RegistrarFallo(ctx context.Context, contexto, id string) (EstadoIntentos, error) {
	segundos := int(l.bloqueo.Seconds())
	res, err := scriptFallo.Run(ctx, l.client,
		[]string{claveIntentos(contexto, id), claveBloqueo(contexto, id)},
		l.maxIntentos, segundos, segundos).Slice()
	if err != nil {
		return EstadoIntentos{}, fmt.Errorf("registrar intento fallido: %w", err)
	}
	if len(res) != 2 {
		return EstadoIntentos{}, fmt.Errorf("respuesta inesperada del script de intentos")
	}
	n, _ := res[0].(int64)
	bloqueado, _ := res[1].(int64)
	if bloqueado == 1 {
		return EstadoIntentos{Bloqueado: true, SegundosBloqueo: segundos}, nil
	}
	restantes := l.maxIntentos - int(n)
	if restantes < 0 {
		restantes = 0
	}
	return EstadoIntentos{Restantes: restantes}, nil
}

func (l *limitadorRedis) Reiniciar(ctx context.Context, contexto, id string) error {
	if err := l.client.Del(ctx, claveIntentos(contexto, id)).Err(); err != nil {
		return fmt.Errorf("reiniciar intentos: %w", err)
	}
	return nil
}
