import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators'
import { apiSocket } from '@/api/api'

export interface User {
  name: string | null
  handle: string | null
}

export interface StartSessionRequest {
  action?: 'newSession'
  facilitator: User
  facilitatorParticipating: boolean
}

export interface PointingSession {
  isFacilitator: boolean
  facilitatorSessionKey?: string
  sessionId: string
  facilitator: User
  participants: Array<User>
}

export interface PointingSessionState {
  currentSession: PointingSession | null
  knownSessions: Array<PointingSession>
}

@Module
export class PointingSessionStore extends VuexModule<PointingSessionState> {
  static ACTION_INITIALIZE = 'initialize'

  static ACTION_BEGIN_SESSION = 'beginSession'

  static MUTATION_ADD_SESSION = 'addSession'

  static MUTATION_SET_SESSION = 'setCurrentSession'

  currentSession: PointingSession | undefined

  knownSessions: Array<PointingSession> = []

  @Mutation
  addSession(session: PointingSession) {
    this.knownSessions.push(session)
  }

  @Mutation
  setCurrentSession(sessionId: string | null) {
    this.currentSession = this.knownSessions.find((sess) => {
      return sess.sessionId === sessionId
    })
  }

  @Action
  beginSession(request: StartSessionRequest) {
    request.action = 'newSession'
    apiSocket().send(JSON.stringify(request))
  }
}
