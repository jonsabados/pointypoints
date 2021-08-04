import Vue from 'vue'

export interface AuthResponse {
  // eslint-disable-next-line
  id_token: string // (WTF google, really, snake case???)
}

export interface GoogleUser {
  getAuthResponse(): AuthResponse
  isSignedIn(): boolean
}

export interface AuthListener {
  (currentUser: GoogleUser): void
}

// vue-google-auth does not play well with typescript so we will just wrap it all so other things can be tested
export async function isSignedIn(): Promise<boolean> {
  // @ts-ignore
  const auth2 = await Vue.GoogleAuth
  return auth2.isSignedIn.get()
}

export async function currentUser(): Promise<GoogleUser> {
  // @ts-ignore
  const auth2 = await Vue.GoogleAuth
  return auth2.currentUser.get()
}

export async function listenForUser(listener: AuthListener) {
  // @ts-ignore
  const auth2 = await Vue.GoogleAuth
  auth2.currentUser.listen(listener)
}
