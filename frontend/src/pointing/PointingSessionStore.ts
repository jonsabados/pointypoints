import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators'
import { AppStore } from '@/app/AppStore'
import { User } from '@/user/user'

const EVENT_TYPE_SESSION_CREATED = 'SESSION_CREATED'
const FACILITATOR_SESSION_LOADED = 'FACILITATOR_SESSION_LOADED'
const SESSION_LOADED = 'SESSION_LOADED'
const SESSION_UPDATED = 'SESSION_UPDATED'
const PING = 'PING'

export interface JoinSessionRequest {
  action?: 'joinSession'
  sessionId: string
  user: User
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

export interface VoteRequest {
  action?: 'vote'
  sessionId: string
  vote: string
}

export interface ClearVotesRequest {
  action?: 'clearVotes'
  sessionId: string
  facilitatorSessionKey: string
}

export interface PointingSession {
  isFacilitator: boolean
  facilitatorSessionKey?: string
  facilitatorPoints: boolean
  sessionId: string
  facilitator: User
  participants: Array<User>
  votesShown: boolean
}

export interface PointingSessionState {
  connectionId: string | null
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
    facilitatorPoints: view.facilitatorPoints,
    votesShown: view.votesShown
  }
}

@Module
export class PointingSessionStore extends VuexModule<PointingSessionState> {
  static ACTION_INITIALIZE = 'initialize'
  static ACTION_END_SESSION = 'endSession'
  static ACTION_LOAD_FACILITATOR_SESSION = 'loadFacilitatorSession'
  static ACTION_SET_FACILITATOR_SESSION = 'setFacilitatorSession'
  static ACTION_LOAD_SESSION = 'loadSession'

  static MUTATION_SET_ACTIVE_SESSION = 'setActiveSession'
  static MUTATION_END_SESSION = 'clearSession'
  static MUTATION_SESSION_ADDED = 'sessionAdded'
  static MUTATION_SET_SESSIONS = 'setSessions'
  static MUTATION_SET_CONNECTION_ID = 'setConnectionId'

  connectionId: string | null = null

  sessionActive: boolean = false

  // this sucks and needs to go away but no time to do it cleanly - its the facilitator session
  currentSession: PointingSession | undefined

  knownSessions: Array<PointingSession> = []

  socket: WebSocket = new WebSocket(`${process.env['VUE_APP_POINTING_SOCKET_URL']}/`)

  @Mutation
  setConnectionId(connectionId: string) {
    this.connectionId = connectionId
  }

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

  @Mutation
  setSessions(sessions: Array<PointingSession>) {
    this.knownSessions = sessions
    if (this.currentSession) {
      sessions.forEach((s) => {
        if (this.currentSession && s.sessionId === this.currentSession.sessionId) {
          // this also sucks but there is something with updates not being seen that I don't understand currently and don't have time to figure out
          this.currentSession.participants = s.participants
          this.currentSession.votesShown = s.votesShown
          this.currentSession.facilitator = s.facilitator
        }
      })
    }
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
        case SESSION_UPDATED: {
          const newSessions: PointingSession[] = []
          this.knownSessions.forEach((s) => {
            if (s.sessionId === eventData.body.sessionId) {
              newSessions.push(eventData.body)
            } else {
              newSessions.push(s)
            }
          })
          this.context.commit(PointingSessionStore.MUTATION_SET_SESSIONS, newSessions)
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
  setFacilitatorSession(session: PointingSession) {
    const view = convertFacilitatorSession(session)
    this.context.commit(PointingSessionStore.MUTATION_SESSION_ADDED, view)
    this.context.commit(PointingSessionStore.MUTATION_SET_ACTIVE_SESSION, view)
  }

  @Action
  loadSession(request: LoadSessionRequest) {
    request.action = 'loadSession'
    sendMessage(this.socket, request)
  }

  @Action
  endSession() {
    this.context.commit(PointingSessionStore.MUTATION_END_SESSION)
  }
}
