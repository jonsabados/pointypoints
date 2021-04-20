import { v4 as uuidv4 } from 'uuid'

export interface User {
  userId?: string
  connectionId: string
  name: string
  handle?: string
  currentVote?: string
}

export function newUser(name: string, handle?: string):User {
  const userId = uuidv4()
  return { userId, name, handle, connectionId: '' }
}
