import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators'
import { currentUser, GoogleUser, isSignedIn, listenForUser } from '@/profile/google'
import { getProfile, Profile } from '@/pointing/pointing'
import { AppStore } from '@/app/AppStore'

export interface ProfileState {
  isReady: boolean
  signedIn: boolean
  authToken: string
  profile: Profile | undefined
}

@Module
export class ProfileStore extends VuexModule<ProfileState> {
  static ACTION_FETCH_PROFILE = 'fetchProfile'

  isReady: boolean = false
  signedIn: boolean = false
  authToken: string = ''
  profile: Profile | undefined

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
  setProfile(profile: Profile | undefined) {
    this.profile = profile
  }

  @Mutation
  markReady() {
    this.isReady = true
  }

  @Action
  async fetchProfile() {
    try {
      const profile = await getProfile(this.authToken)
      this.context.commit('setProfile', profile)
    } catch (e) {
      await this.context.dispatch(AppStore.ACTION_REGISTER_REMOTE_ERROR, e)
    }
  }

  @Action
  async initialize() {
    // missing await is very intentional, don't wanna block
    listenForUser((user) => {
      this.context.commit('setProfile', undefined)
      this.context.commit('setGoogleUser', user)
      if (user.isSignedIn()) {
        this.context.dispatch(ProfileStore.ACTION_FETCH_PROFILE)
      }
    })

    const loggedIn = await isSignedIn()
    this.context.commit('markReady')

    if (loggedIn) {
      const user = await currentUser()
      this.context.commit('setGoogleUser', user)
      await this.context.dispatch(ProfileStore.ACTION_FETCH_PROFILE)
    }
  }
}
