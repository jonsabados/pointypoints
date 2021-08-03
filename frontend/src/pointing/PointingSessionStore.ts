import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators'
import { AppStore } from '@/app/AppStore'
import { User } from '@/user/user'

const SESSION_UPDATED = 'SESSION_UPDATED'
const PING = 'PING'

export interface PointingSession {
  facilitatorPoints: boolean
  sessionId: string
  facilitator: User
  participants: Array<User>
  votesShown: boolean
}

export interface PointingSessionState {
  facilitating: boolean
  connectionId: string | null
  sessionId: string | null
  currentSession: PointingSession | null
}

function sendMessage(socket: WebSocket, message: any) {
  const doIt = () => {
    socket.send(JSON.stringify(message))
  }
  if (socket.readyState === 1) {
    doIt()
  } else {
    const orig = socket.onopen
    socket.onopen = (ev) => {
      if (orig) {
        // @ts-ignore
        orig(ev)
      }
      doIt()
    }
  }
}

@Module
export class PointingSessionStore extends VuexModule<PointingSessionState> {
  static ACTION_INITIALIZE = 'initialize'
  static ACTION_END_SESSION = 'endSession'

  static MUTATION_SET_FACILITATING = 'setFacilitating'
  static MUTATION_SET_SESSION = 'setSession'
  static MUTATION_SET_SESSION_ID = 'setSessionId'
  static MUTATION_END_SESSION = 'clearSession'
  static MUTATION_SET_CONNECTION_ID = 'setConnectionId'

  facilitating: boolean = false

  connectionId: string | null = null

  sessionId: string | null = null

  currentSession: PointingSession | null = null

  socket: WebSocket = new WebSocket(`${process.env['VUE_APP_POINTING_SOCKET_URL']}/`)

  @Mutation
  setFacilitating(facilitating: boolean) {
    this.facilitating = facilitating
  }

  @Mutation
  setConnectionId(connectionId: string) {
    this.connectionId = connectionId
  }

  @Mutation
  clearSession() {
    this.currentSession = null
    this.sessionId = null
  }

  @Mutation
  setSession(session: PointingSession) {
    this.currentSession = session
  }

  @Mutation
  setSessionId(sessionId: string) {
    this.sessionId = sessionId
  }

  @Action
  initialize() {
    const ping = () => {
      sendMessage(this.socket, {
        action: 'ping'
      })
    }
    setInterval(ping, 30000)
    ping()
    this.socket.onerror = (ev) => {
      this.context.dispatch(AppStore.ACTION_REGISTER_REMOTE_ERROR, ev)
    }
    this.socket.onmessage = (ev) => {
      const eventData = JSON.parse(ev.data)
      if (!eventData.type) {
        console.log('event does not match expected interface')
        console.log(eventData)
        return
      }
      switch (eventData.type) {
        case SESSION_UPDATED: {
          const session = eventData.body as PointingSession
          if (session.sessionId === this.sessionId) {
            this.context.commit(PointingSessionStore.MUTATION_SET_SESSION, session)
          }
          break
        }
        case PING: {
          this.context.commit(PointingSessionStore.MUTATION_SET_CONNECTION_ID, eventData.body.connectionId)
          break
        }
        default:
          console.log('unknown event')
          console.log(eventData)
      }
    }
  }

  @Action
  endSession() {
    this.context.commit(PointingSessionStore.MUTATION_END_SESSION)
  }
}
