import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators'
import { currentUser, GoogleUser, isSignedIn, listenForUser } from '@/profile/google'

export interface ProfileState {
  isReady: boolean
  signedIn: boolean
  authToken: string
}

@Module
export class ProfileStore extends VuexModule<ProfileState> {
  isReady: boolean = false

  signedIn: boolean = false

  authToken: string = ''

  @Mutation
  setGoogleUser(user: GoogleUser) {
    if (user.isSignedIn()) {
      this.signedIn = true
      this.authToken = `Bearer ${user.getAuthResponse().id_token}`
    } else {
      this.signedIn = false
      this.authToken = ''
    }
  }

  @Mutation
  markReady() {
    this.isReady = true
  }

  @Action
  async initialize() {
    // missing await is very intentional, don't wanna block
    listenForUser((user) => {
      this.context.commit('setGoogleUser', user)
    })

    const loggedIn = await isSignedIn()
    this.context.commit('markReady')

    if (loggedIn) {
      const user = await currentUser()
      this.context.commit('setGoogleUser', user)
      // this.context.dispatch(UserStore.ACTION_FETCH_SELF)
    }
  }
}
