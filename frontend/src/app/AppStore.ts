import { Module, VuexModule, Mutation, Action } from 'vuex-module-decorators'

export interface AppState {
  blockingOpRunning: boolean
  errorToAck: any
}

@Module
export class AppStore extends VuexModule<AppState> {
  static ACTION_REGISTER_REMOTE_ERROR = 'registerRemoteError'
  static ACTION_ACK_REMOTE_ERROR = 'ackRemoteError'
  static MUTATION_SET_ERROR_TO_ACK = 'setErrorToAck'

  errorToAck: any | null = null

  waitingOnServer: boolean = false

  @Mutation
  setErrorToAck(error: any | null) {
    this.errorToAck = error
  }

  @Mutation
  setWaitingOnServer(waitingOnServer: boolean) {
    this.waitingOnServer = waitingOnServer
  }

  @Action
  registerRemoteError(error: any) {
    // eslint-disable-next-line
    console.error(error)
    this.context.commit(AppStore.MUTATION_SET_ERROR_TO_ACK, error)
  }

  @Action
  ackRemoteError() {
    this.context.commit(AppStore.MUTATION_SET_ERROR_TO_ACK, null)
  }
}
