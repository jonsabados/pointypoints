import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators'
import { AppStore } from '@/app/AppStore'

const EVENT_TYPE_SESSION_CREATED = 'SESSION_CREATED'
const FACILITATOR_SESSION_LOADED = 'FACILITATOR_SESSION_LOADED'
const SESSION_LOADED = 'SESSION_LOADED'

export interface User {
  name: string | null
  handle: string | null
  currentVote?: number
}

export interface LoadFacilitatorSessionRequest {
  action?: 'loadFacilitatorSession'
  sessionId: string
  facilitatorSessionKey: string
  markActive: boolean
}

export interface LoadSessionRequest {
  action?: 'loadSession'
  sessionId: string
  markActive: boolean
}

export interface StartSessionRequest {
  action?: 'newSession'
  facilitator: User
  facilitatorParticipating: boolean
}

export interface PointingSession {
  isFacilitator: boolean
  facilitatorSessionKey?: string
  facilitatorPoints: boolean
  sessionId: string
  facilitator: User
  participants: Array<User>
}

export interface PointingSessionState {
  sessionActive: boolean
  currentSession: PointingSession | null
  knownSessions: Array<PointingSession>
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

function convertFacilitatorSession(view: any): PointingSession {
  return {
    isFacilitator: true,
    facilitatorSessionKey: view.facilitatorSessionKey,
    sessionId: view.sessionId,
    facilitator: view.facilitator,
    participants: view.participants,
    facilitatorPoints: view.facilitatorPoints
  }
}

@Module
export class PointingSessionStore extends VuexModule<PointingSessionState> {
  static ACTION_INITIALIZE = 'initialize'
  static ACTION_BEGIN_SESSION = 'beginSession'
  static ACTION_END_SESSION = 'endSession'
  static ACTION_LOAD_FACILITATOR_SESSION = 'loadFacilitatorSession'
  static ACTION_LOAD_SESSION = 'loadSession'

  static MUTATION_SET_ACTIVE_SESSION = 'setActiveSession'
  static MUTATION_END_SESSION = 'clearSession'
  static MUTATION_SESSION_ADDED = 'sessionAdded'

  sessionActive: boolean = false

  currentSession: PointingSession | undefined

  knownSessions: Array<PointingSession> = []

  socket: WebSocket = new WebSocket(`${process.env['VUE_APP_POINTING_SOCKET_URL']}/`)

  @Mutation
  clearSession() {
    this.sessionActive = false
    this.currentSession = undefined
  }

  @Mutation
  setActiveSession(session: PointingSession) {
    this.currentSession = session
    this.sessionActive = true
  }

  @Mutation
  sessionAdded(session: PointingSession) {
    this.knownSessions.push(session)
  }

  @Action
  initialize() {
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
        case EVENT_TYPE_SESSION_CREATED: {
          const session = convertFacilitatorSession(eventData.body)
          this.context.commit(PointingSessionStore.MUTATION_SESSION_ADDED, session)
          this.context.commit(PointingSessionStore.MUTATION_SET_ACTIVE_SESSION, session)
          break
        }
        case FACILITATOR_SESSION_LOADED: {
          const session = convertFacilitatorSession(eventData.body.session)
          this.context.commit(PointingSessionStore.MUTATION_SESSION_ADDED, session)
          if (eventData.body.markActive) {
            this.context.commit(PointingSessionStore.MUTATION_SET_ACTIVE_SESSION, session)
          }
          break
        }
        case SESSION_LOADED: {
          this.context.commit(PointingSessionStore.MUTATION_SESSION_ADDED, eventData.body)
          break
        }
        default:
          console.log('unknown event')
          console.log(eventData)
      }
    }
  }

  @Action
  loadFacilitatorSession(request: LoadFacilitatorSessionRequest) {
    if (this.knownSessions.find((s) => {
      return s.sessionId === request.sessionId
    })) {
      return
    }
    request.action = 'loadFacilitatorSession'
    sendMessage(this.socket, request)
  }

  @Action
  loadSession(request: LoadSessionRequest) {
    request.action = 'loadSession'
    sendMessage(this.socket, request)
  }

  @Action
  beginSession(request: StartSessionRequest) {
    request.action = 'newSession'
    this.socket.send(JSON.stringify(request))
  }

  @Action
  endSession() {
    this.context.commit(PointingSessionStore.MUTATION_END_SESSION)
  }
}
