'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { Button, Badge, Card } from '@/components/ui'
import { UserShell } from '@/components/layout/UserShell'
import { UserRouteLink } from '@/components/layout/UserRouteLink'
import { apiGet } from '@/lib/api'
import { useAppStore } from '@/lib/store'
import { formatDateTime, getOrderStatus, copyToClipboard, cn } from '@/lib/utils'
import { getCachedUserInfo, getUserInfo, updateCachedUserInfo } from '@/lib/userData'
import type { TwoFAStatus, UserInfo } from '@/types/user'
import {
  BindEmailModal,
  ChangeEmailModal,
  ChangePasswordModal,
  Setup2FAModal,
  Disable2FAModal,
  ChangeMethodModal,
} from './modals/index'
import {
  KamisTab,
  WalletTab,
} from './tabs'

/**
 * 订单接口
 */
interface Order {
  id: number
  order_no: string
  product_name: string
  price: number
  status: number
  kami_code: string
  created_at: string
}

/**
 * 用户中心视图。
 * 由用户端主入口和旧路径直达共同复用。
 */
export function UserCenterView() {
  const router = useRouter()
  const { user, setUser, twoFAStatus, setTwoFAStatus } = useAppStore()
  const [activeTab, setActiveTab] = useState('profile')
  const [loading, setLoading] = useState(() => !getCachedUserInfo() && !user)
  const [orders, setOrders] = useState<Order[]>([])

  // 加载用户信息
  useEffect(() => {
    const loadUserInfo = async () => {
      const userInfo = await getUserInfo()
      if (userInfo) {
        setUser(userInfo)
        updateCachedUserInfo(userInfo)
        await load2FAStatus()
      } else {
        router.push('/login/')
      }
      setLoading(false)
    }
    loadUserInfo()
  }, [setUser, router])

  // 加载 2FA 状态
  const load2FAStatus = async () => {
    const res = await apiGet<{ enabled: boolean; has_totp: boolean; prefer_email_auth: boolean }>('/api/user/2fa/status')
    if (res.success) {
      setTwoFAStatus({
        enabled: res.enabled,
        has_totp: res.has_totp,
        prefer_email_auth: res.prefer_email_auth,
      })
    }
  }

  // 加载订单
  const loadOrders = async () => {
    const res = await apiGet<{ orders: Order[] }>('/api/user/orders')
    if (res.success && res.orders) {
      setOrders(res.orders)
    }
  }

  // 切换标签页
  const handleTabChange = (tab: string) => {
    setActiveTab(tab)
    if (tab === 'orders') {
      loadOrders()
    }
  }

  // 标签页配置
  const tabs = [
    { id: 'profile', label: '个人信息', icon: 'fa-user' },
    { id: 'orders', label: '我的订单', icon: 'fa-bag-shopping' },
    { id: 'kamis', label: '我的卡密', icon: 'fa-key' },
    { id: 'wallet', label: '我的钱包', icon: 'fa-wallet' },
    { id: 'security', label: '安全设置', icon: 'fa-shield-halved' },
  ]

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <i className="fas fa-spinner fa-spin text-4xl text-primary-400" />
      </div>
    )
  }

  return (
    <>
      <main className="flex-1 py-8 px-4">
        <div className="max-w-5xl mx-auto">
          {/* 标签页导航 - 支持滚动，隐藏滚动条 */}
          <div className="overflow-x-auto -mx-4 px-4 mb-6 scrollbar-hide">
            <div className="flex border-b border-dark-700/50 min-w-max gap-1">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => handleTabChange(tab.id)}
                  className={cn(
                    'px-4 py-3 text-sm font-medium whitespace-nowrap transition-colors rounded-t-lg',
                    activeTab === tab.id 
                      ? 'text-primary-400 bg-primary-500/10 border-b-2 border-primary-500' 
                      : 'text-dark-400 hover:text-dark-200 hover:bg-dark-700/30'
                  )}
                >
                  <i className={`fas ${tab.icon} mr-2`} />
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* 个人信息 */}
          {activeTab === 'profile' && user && (
            <ProfileTab user={user} onUpdate={() => {
              apiGet<{ user: typeof user }>('/api/user/info').then(res => {
                if (res.success && res.user) {
                  setUser(res.user)
                  updateCachedUserInfo(res.user)
                }
              })
            }} />
          )}

          {/* 我的订单 */}
          {activeTab === 'orders' && <OrdersTab orders={orders} />}

          {/* 我的卡密 */}
          {activeTab === 'kamis' && <KamisTab />}

          {/* 我的钱包 */}
          {activeTab === 'wallet' && <WalletTab />}

          {/* 安全设置 */}
          {activeTab === 'security' && user && twoFAStatus && (
            <SecurityTab
              user={user}
              twoFAStatus={twoFAStatus}
              onUpdate={load2FAStatus}
            />
          )}
        </div>
      </main>
    </>
  )
}

/**
 * 用户中心旧路径直达入口。
 */
export default function UserCenterPage() {
  return (
    <UserShell>
      <UserCenterView />
    </UserShell>
  )
}

/**
 * 个人信息标签页
 */
function ProfileTab({ user, onUpdate }: { user: UserInfo; onUpdate: () => void }) {
  const [showBindEmail, setShowBindEmail] = useState(false)
  const [showChangeEmail, setShowChangeEmail] = useState(false)

  return (
    <div>
      <Card title="基本信息" icon={<i className="fas fa-user" />}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <InfoItem label="用户名" value={user.username} />
          <InfoItem
            label="邮箱"
            value={
              user.email ? (
                <div className="flex items-center gap-2 flex-wrap">
                  <span>{user.email}</span>
                  <Badge variant={user.email_verified ? 'success' : 'warning'}>
                    {user.email_verified ? '已验证' : '未验证'}
                  </Badge>
                  <button
                    onClick={() => setShowChangeEmail(true)}
                    className="text-primary-400 hover:text-primary-300 text-sm"
                  >
                    更换
                  </button>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <span className="text-dark-500">未绑定</span>
                  <button
                    onClick={() => setShowBindEmail(true)}
                    className="text-primary-400 hover:text-primary-300 text-sm"
                  >
                    绑定邮箱
                  </button>
                </div>
              )
            }
          />
          <InfoItem label="手机" value={user.phone || '未设置'} />
          <InfoItem label="注册时间" value={formatDateTime(user.created_at)} />
        </div>
      </Card>

      {/* 绑定邮箱弹窗 */}
      <BindEmailModal
        isOpen={showBindEmail}
        onClose={() => setShowBindEmail(false)}
        onSuccess={onUpdate}
      />

      {/* 更换邮箱弹窗 */}
      <ChangeEmailModal
        isOpen={showChangeEmail}
        onClose={() => setShowChangeEmail(false)}
        currentEmail={user.email || ''}
        onSuccess={onUpdate}
      />
    </div>
  )
}

/**
 * 信息项组件
 */
function InfoItem({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="bg-dark-700/30 rounded-xl p-4">
      <div className="text-dark-500 text-sm mb-1">{label}</div>
      <div className="text-dark-100">{value}</div>
    </div>
  )
}

/**
 * 订单标签页
 */
function OrdersTab({ orders }: { orders: Order[] }) {
  const handleCopyKami = async (code: string) => {
    const success = await copyToClipboard(code)
    if (success) toast.success('已复制到剪贴板')
  }

  return (
    <div>
      <Card title="订单列表" icon={<i className="fas fa-shopping-bag" />}>
        {orders.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-5xl mb-4">📦</div>
            <p className="text-dark-400">暂无订单</p>
          </div>
        ) : (
          <div className="space-y-4">
            {orders.map((order) => {
              const status = getOrderStatus(order.status)
              return (
                <div
                  key={order.id}
                  className="bg-dark-700/30 rounded-xl p-4 border border-dark-600/50"
                >
                  <div className="flex justify-between items-start mb-3">
                    <span className="text-dark-500 text-sm font-mono">
                      订单号: {order.order_no}
                    </span>
                    <Badge variant={status.variant}>{status.text}</Badge>
                  </div>
                  <div className="text-dark-300 text-sm mb-2">
                    商品: {order.product_name} | 金额: ¥{order.price.toFixed(2)} |{' '}
                    {formatDateTime(order.created_at)}
                  </div>
                  {order.kami_code && (
                    <div className="mt-3 bg-dark-800/50 rounded-lg p-3 flex items-center justify-between">
                      <span className="font-mono text-primary-400 break-all">
                        {order.kami_code}
                      </span>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => handleCopyKami(order.kami_code)}
                      >
                        <i className="fas fa-copy" />
                      </Button>
                    </div>
                  )}
                  <div className="mt-3 flex justify-end">
                    <UserRouteLink
                      view="order-detail"
                      params={{ order_no: order.order_no }}
                      className="text-primary-400 hover:text-primary-300 text-sm flex items-center gap-1"
                    >
                      查看详情
                      <i className="fas fa-chevron-right text-xs" />
                    </UserRouteLink>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </Card>
    </div>
  )
}

/**
 * 安全设置标签页
 */
function SecurityTab({
  user,
  twoFAStatus,
  onUpdate,
}: {
  user: UserInfo
  twoFAStatus: TwoFAStatus
  onUpdate: () => void
}) {
  const [showChangePassword, setShowChangePassword] = useState(false)
  const [showSetup2FA, setShowSetup2FA] = useState(false)
  const [showDisable2FA, setShowDisable2FA] = useState(false)
  const [showChangeMethod, setShowChangeMethod] = useState(false)

  const isUsingTOTP = twoFAStatus.has_totp && !twoFAStatus.prefer_email_auth

  // 检查是否可以开启2FA
  const canEnable2FA = user.email && user.email_verified

  const handleSetup2FA = () => {
    if (!user.email) {
      toast.error('请先绑定邮箱')
      return
    }
    if (!user.email_verified) {
      toast.error('请先验证邮箱')
      return
    }
    setShowSetup2FA(true)
  }

  return (
    <div className="space-y-6">
      {/* 修改密码 */}
      <Card title="修改密码" icon={<i className="fas fa-key" />}>
        <div className="security-card">
          <div>
            <h4 className="text-dark-100 font-medium">账号密码</h4>
            <p className="text-dark-500 text-sm">定期修改密码可以提高账号安全性</p>
          </div>
          <Button size="sm" onClick={() => setShowChangePassword(true)}>
            修改密码
          </Button>
        </div>
      </Card>

      {/* 两步验证 */}
      <Card title="两步验证" icon={<i className="fas fa-shield-alt" />}>
        <div className="security-card mb-4">
          <div>
            <h4 className="text-dark-100 font-medium">两步验证</h4>
            <p className="text-dark-500 text-sm">
              {twoFAStatus.enabled
                ? isUsingTOTP
                  ? '当前使用动态口令方式'
                  : '当前使用邮箱验证码方式'
                : '登录时需要额外验证，提高账号安全性'}
            </p>
          </div>
          {twoFAStatus.enabled ? (
            <Button size="sm" variant="danger" onClick={() => setShowDisable2FA(true)}>
              关闭
            </Button>
          ) : (
            <Button size="sm" onClick={handleSetup2FA} disabled={!canEnable2FA}>
              开启
            </Button>
          )}
        </div>

        {twoFAStatus.enabled && (
          <div className="security-card">
            <div>
              <h4 className="text-dark-100 font-medium">
                {isUsingTOTP ? '切换到邮箱验证' : '切换到动态口令'}
              </h4>
              <p className="text-dark-500 text-sm">
                {isUsingTOTP ? '使用邮箱接收验证码' : '使用验证器APP生成动态口令'}
              </p>
            </div>
            <Button size="sm" variant="secondary" onClick={() => setShowChangeMethod(true)}>
              更改
            </Button>
          </div>
        )}
      </Card>

      {/* 修改密码弹窗 */}
      <ChangePasswordModal
        isOpen={showChangePassword}
        onClose={() => setShowChangePassword(false)}
        userEmail={user.email || ''}
        twoFAStatus={twoFAStatus}
      />

      {/* 开启两步验证弹窗 */}
      <Setup2FAModal
        isOpen={showSetup2FA}
        onClose={() => setShowSetup2FA(false)}
        userEmail={user.email || ''}
        onSuccess={onUpdate}
      />

      {/* 关闭两步验证弹窗 */}
      <Disable2FAModal
        isOpen={showDisable2FA}
        onClose={() => setShowDisable2FA(false)}
        userEmail={user.email || ''}
        twoFAStatus={twoFAStatus}
        onSuccess={onUpdate}
      />

      {/* 更改验证方式弹窗 */}
      <ChangeMethodModal
        isOpen={showChangeMethod}
        onClose={() => setShowChangeMethod(false)}
        userEmail={user.email || ''}
        isUsingTOTP={isUsingTOTP}
        onSuccess={onUpdate}
      />
    </div>
  )
}
