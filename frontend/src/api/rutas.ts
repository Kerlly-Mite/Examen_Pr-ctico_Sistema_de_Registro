export const RUTAS = {
  verificarIdentidad: '/api/auth/verificar-identidad',
  registro: '/api/auth/registro',
  login: '/api/auth/login',
  me: '/api/auth/me',
  recuperarSolicitar: '/api/auth/recuperar/solicitar',
  recuperarRestablecer: '/api/auth/recuperar/restablecer',
  paises: '/api/catalogos/paises',
  provincias: (paisId: number) => `/api/catalogos/paises/${paisId}/provincias`,
  cantones: (provinciaId: number) => `/api/catalogos/provincias/${provinciaId}/cantones`,
} as const