export interface UserInfo {
  id: number
  username: string
  email: string
  email_verified: boolean
  phone: string
  created_at: string
}

export interface TwoFAStatus {
  enabled: boolean
  has_totp: boolean
  prefer_email_auth: boolean
}
