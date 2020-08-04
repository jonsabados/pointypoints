const socket = new WebSocket(`${process.env['VUE_APP_POINTING_SOCKET_URL']}/`)

export function apiSocket(): WebSocket {
  return socket
}
