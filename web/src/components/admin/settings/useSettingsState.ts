'use client'

import { useState, useEffect, useCallback } from 'react'
import { apiGet, apiPost } from '@/lib/api'
import { Settings } from '../types'

/**
 * 基本设置 / 登录设置 / 安全设置 共享的数据状态。
 *
 * 后端 AdminSaveSecuritySettings 接口为全量覆盖式保存：
 * 先读取当前完整配置，再用请求体字段无条件覆盖 enable_2fa / totp_secret 等。
 * 因此登录设置与安全设置必须各自独立保存时，把不属于本页的字段用当前值一并带上，
 * 否则会把另一页的配置清零。本 hook 统一维护一份完整状态供三个子页面共享。
 */
export interface SettingsState {
  loading: boolean
  basicForm: { system_title: string; admin_suffix: string; server_port: string }
  securityForm: {
    enable_login: boolean
    enable_captcha: boolean
    admin_username: string
    admin_password: string
    enable_2fa: boolean
    totp_secret: string
    enable_session_timeout: boolean
    session_timeout: number
    user_allow_register: boolean
    user_enable_captcha: boolean
    user_enable_2fa: boolean
    user_require_email_verification: boolean
    user_enable_session_timeout: boolean
    user_session_timeout: number
  }
  loadSettings: () => Promise<void>
  setBasicForm: React.Dispatch<React.SetStateAction<SettingsState['basicForm']>>
  setSecurityForm: React.Dispatch<React.SetStateAction<SettingsState['securityForm']>>
  /** 保存基本设置（系统标题、后台后缀、端口） */
  saveBasic: () => Promise<boolean>
  /** 保存登录设置与安全设置（全量提交，避免互相覆盖） */
  saveSecurity: () => Promise<boolean>
}

export function useSettingsState(): SettingsState {
  const [loading, setLoading] = useState(true)
  const [basicForm, setBasicForm] = useState({ system_title: '', admin_suffix: 'manage', server_port: '8080' })
  const [securityForm, setSecurityForm] = useState({
    enable_login: true,
    enable_captcha: true,
    admin_username: 'admin',
    admin_password: '',
    enable_2fa: false,
    totp_secret: '',
    enable_session_timeout: true,
    session_timeout: 60,
    user_allow_register: true,
    user_enable_captcha: true,
    user_enable_2fa: true,
    user_require_email_verification: false,
    user_enable_session_timeout: true,
    user_session_timeout: 120,
  })

  const loadSettings = useCallback(async () => {
    const res = await apiGet<{ settings: Settings }>('/api/admin/settings')
    if (res.success && res.settings) {
      setBasicForm({
        system_title: res.settings.system_title || '',
        admin_suffix: res.settings.admin_suffix || 'manage',
        server_port: String(res.settings.server_port || 8080)
      })
      setSecurityForm({
        enable_login: res.settings.enable_login,
        enable_captcha: res.settings.enable_captcha ?? true,
        admin_username: res.settings.admin_username || 'admin',
        admin_password: '',
        enable_2fa: res.settings.enable_2fa,
        totp_secret: res.settings.totp_secret || '',
        enable_session_timeout: res.settings.enable_session_timeout ?? true,
        session_timeout: res.settings.session_timeout || 60,
        user_allow_register: res.settings.user_allow_register ?? true,
        user_enable_captcha: res.settings.user_enable_captcha ?? true,
        user_enable_2fa: res.settings.user_enable_2fa ?? true,
        user_require_email_verification: res.settings.user_require_email_verification ?? false,
        user_enable_session_timeout: res.settings.user_enable_session_timeout ?? true,
        user_session_timeout: res.settings.user_session_timeout || 120,
      })
    }
    setLoading(false)
  }, [])

  useEffect(() => { loadSettings() }, [loadSettings])

  const saveBasic = useCallback(async () => {
    const suffix = basicForm.admin_suffix.trim()
    if (suffix && !/^[a-zA-Z0-9_-]+$/.test(suffix)) {
      return false
    }
    const port = parseInt(basicForm.server_port) || 8080
    if (port < 1 || port > 65535) {
      return false
    }
    const res = await apiPost('/api/admin/settings', {
      system_title: basicForm.system_title.trim(),
      admin_suffix: suffix || 'manage',
      server_port: port
    })
    if (res.success) {
      // 重新加载以确保显示最新数据
      await loadSettings()
      return true
    }
    return false
  }, [basicForm, loadSettings])

  const saveSecurity = useCallback(async () => {
    // 全量提交：登录设置与安全设置字段一起发送，避免后端覆盖式保存导致另一页配置丢失
    const data: Record<string, unknown> = {
      enable_login: securityForm.enable_login,
      enable_captcha: securityForm.enable_captcha,
      admin_username: securityForm.admin_username.trim() || 'admin',
      enable_2fa: securityForm.enable_2fa,
      totp_secret: securityForm.totp_secret,
      enable_session_timeout: securityForm.enable_session_timeout,
      session_timeout: securityForm.session_timeout,
      user_allow_register: securityForm.user_allow_register,
      user_enable_captcha: securityForm.user_enable_captcha,
      user_enable_2fa: securityForm.user_enable_2fa,
      user_require_email_verification: securityForm.user_require_email_verification,
      user_enable_session_timeout: securityForm.user_enable_session_timeout,
      user_session_timeout: securityForm.user_session_timeout,
    }
    if (securityForm.admin_password) data.admin_password = securityForm.admin_password
    const res = await apiPost('/api/admin/settings/security', data)
    if (res.success) {
      await loadSettings()
      return true
    }
    return false
  }, [securityForm, loadSettings])

  return { loading, basicForm, securityForm, loadSettings, setBasicForm, setSecurityForm, saveBasic, saveSecurity }
}
