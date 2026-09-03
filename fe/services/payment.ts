import { fetchWithAuth } from "./api"
import { apiUrl } from "@/lib/api-base-url"

export interface PaymentMethod {
  code: string
  name: string
  type: string
  min_amount: number
  max_amount: number
  enabled: boolean
  currency?: string
  channel?: string
  description?: string
}

export interface PaymentOrder {
  id: number
  order_id: string
  plat_order_id: string
  amount: number
  status: number
  type: string
  dst_code: string
  pay_url?: string
  created_at: string
  paid_at?: string
}

export interface WalletRecord {
  id: string | number
  record_type?: "deposit" | "withdraw" | "bonus"
  type?: number
  title: string
  order_id?: string
  reference_id?: string
  amount: number
  currency: string
  status: number
  dst_code?: string
  remark?: string
  created_at: string
}

export interface CreatePaymentRequest {
  amount: number
  type: "channel"
  dst_code?: string
  callback_url?: string
  currency?: string
  channel?: string
}

export interface CreatePaymentResponse {
  order_id: string
  pay_url: string
  amount: number
  status: number
  expired_at?: string
}

export interface TelegramStarsOrderResponse {
  order_id: string
  invoice_url: string
  amount: number
  stars_amount: number
  status: number
}

export interface CreateWithdrawRequest {
  amount: number
  type: "bankcard" | "ewallet"
  dst_code: string
  account: string
  account_name: string
  phone?: string
  email?: string
  address?: string
  currency?: string
  channel?: string
  bank_code?: string
}

export interface WithdrawOrder {
  id: number
  order_id: string
  amount: number
  status: number
  type: string
  dst_code: string
  created_at: string
}

export interface WithdrawMethod {
  code: string
  name: string
  type: string
  currency?: string
  channel?: string
  min_amount: number
  max_amount: number
  enabled: boolean
  description?: string
}

export interface PaymentMethodsResponse {
  methods: PaymentMethod[]
}

export interface WithdrawMethodsResponse {
  methods: WithdrawMethod[]
}

const fallbackPaymentMethods: PaymentMethod[] = [
  {
    code: "TG_STARS",
    name: "Telegram Stars",
    type: "channel",
    min_amount: 1,
    max_amount: 100000,
    enabled: false,
    currency: "XTR",
    channel: "telegram",
    description: "Pay securely with Telegram Stars",
  },
  {
    code: "GCASH_QR",
    name: "GCash QR",
    type: "channel",
    min_amount: 100,
    max_amount: 50000,
    enabled: true,
    currency: "PHP",
    channel: "quantixcore",
    description: "Philippines GCash QR payment",
  },
  {
    code: "GCASH_APP",
    name: "GCash App",
    type: "channel",
    min_amount: 100,
    max_amount: 50000,
    enabled: true,
    currency: "PHP",
    channel: "quantixcore",
    description: "Philippines GCash app payment",
  },
  {
    code: "DANA",
    name: "DANA",
    type: "ewallet",
    min_amount: 10000,
    max_amount: 20000000,
    enabled: true,
    currency: "IDR",
    channel: "gateway",
    description: "Indonesia DANA wallet",
  },
  {
    code: "OVO",
    name: "OVO",
    type: "ewallet",
    min_amount: 10000,
    max_amount: 20000000,
    enabled: true,
    currency: "IDR",
    channel: "gateway",
    description: "Indonesia OVO wallet",
  },
  {
    code: "LINKAJA",
    name: "LinkAja",
    type: "ewallet",
    min_amount: 10000,
    max_amount: 20000000,
    enabled: true,
    currency: "IDR",
    channel: "gateway",
    description: "Indonesia LinkAja wallet",
  },
  {
    code: "IDR_QRIS",
    name: "QRIS",
    type: "qris",
    min_amount: 10000,
    max_amount: 10000000,
    enabled: true,
    currency: "IDR",
    channel: "gateway",
    description: "Indonesia QRIS payment",
  },
  {
    code: "IDR_VA",
    name: "Virtual Account",
    type: "bank",
    min_amount: 10000,
    max_amount: 50000000,
    enabled: true,
    currency: "IDR",
    channel: "gateway",
    description: "Indonesia virtual account payment",
  },
  {
    code: "USDT",
    name: "USDT",
    type: "channel",
    min_amount: 77500,
    max_amount: 15500000,
    enabled: true,
    currency: "USDT",
    channel: "tokenpay",
    description: "USDT via TokenPay",
  },
  {
    code: "PAYPAL",
    name: "PayPal",
    type: "channel",
    min_amount: 10000,
    max_amount: 100000000,
    enabled: false,
    currency: "USD",
    channel: "paypal",
    description: "Coming soon",
  },
]

const legacyGatewayWithdrawWalletOptions = [
  { code: "OVO", name: "OVO" },
  { code: "DANA", name: "DANA" },
  { code: "GOPAY", name: "GoPay" },
  { code: "SHOPEEPAY", name: "ShopeePay" },
  { code: "LINKAJA", name: "LinkAja" },
]

const buildLegacyGatewayWithdrawFallbackMethods = (): WithdrawMethod[] => [
  ...legacyGatewayWithdrawWalletOptions.map((option) => ({
    code: option.code,
    name: option.name,
    type: "ewallet",
    min_amount: 10000,
    max_amount: 25000000,
    enabled: true,
    currency: "IDR",
    channel: "gateway",
    description: "Indonesia e-wallet payout",
  })),
]

const fallbackWithdrawMethods: WithdrawMethod[] = [
  {
    code: "GCASH_QR",
    name: "GCash",
    type: "ewallet",
    min_amount: 100,
    max_amount: 50000,
    enabled: true,
    currency: "PHP",
    channel: "quantixcore",
    description: "Philippines GCash wallet",
  },
  ...buildLegacyGatewayWithdrawFallbackMethods(),
]

export const paymentService = {
  async getTonRate(): Promise<number> {
    const response = await fetchWithAuth(apiUrl("/payments/ton/rate"))
    const data = await response.json().catch(() => ({}))
    if (!response.ok) throw new Error(data.error || "Failed to fetch TON rate")
    const rate = Number(data.usd_per_ton)
    if (!(rate > 0)) throw new Error("Invalid TON rate")
    return rate
  },
  async createTonOrder(amount: number): Promise<TonOrderResponse> {
    const response = await fetchWithAuth(apiUrl("/payments/ton/order"), { method:"POST", headers:{"Content-Type":"application/json"}, body:JSON.stringify({amount}) })
    const data = await response.json().catch(() => ({})); if (!response.ok) throw new Error(data.error || "Failed to create TON order"); return data
  },

  async createTelegramStarsOrder(amount: number, initData: string): Promise<TelegramStarsOrderResponse> {
    const response = await fetchWithAuth(apiUrl("/payments/telegram-stars/order"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ amount, init_data: initData }),
    })
    const data = await response.json().catch(() => ({}))
    if (!response.ok) throw new Error(data.error || "Failed to create Telegram Stars invoice")
    return data
  },
  async getPaymentMethods(): Promise<PaymentMethod[]> {
    try {
      const response = await fetch(apiUrl("/payments/methods"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data: PaymentMethodsResponse = await response.json()
      return data.methods
    } catch (error) {
      console.error("Get payment methods failed:", error)
      return fallbackPaymentMethods
    }
  },

  async createPaymentOrder(data: CreatePaymentRequest): Promise<CreatePaymentResponse> {
    try {
      const response = await fetchWithAuth(apiUrl("/payments/order"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      })

      const responseData = await response.json().catch(() => ({}))

      if (!response.ok) {
        const errorMessage = responseData.error || responseData.message || `Request failed (${response.status})`
        throw new Error(errorMessage)
      }

      if (responseData.error) {
        throw new Error(responseData.error)
      }

      return responseData
    } catch (error) {
      console.error("Create payment order failed:", error)
      throw error
    }
  },

  async getPaymentStatus(orderId: string): Promise<PaymentOrder> {
    try {
      const response = await fetchWithAuth(apiUrl(`/payments/order/${orderId}`), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error("Get payment status failed:", error)
      throw error
    }
  },

  async getUserPayments(limit = 20): Promise<WalletRecord[]> {
    try {
      const response = await fetchWithAuth(apiUrl(`/payments/orders?limit=${limit}`), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error("Get user payments failed:", error)
      throw error
    }
  },

  async createWithdrawOrder(data: CreateWithdrawRequest): Promise<WithdrawOrder> {
    try {
      const response = await fetchWithAuth(apiUrl("/withdraw"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      })

      if (!response.ok) {
        const rawError = await response.text().catch(() => "")
        let errorMessage = ""

        if (rawError) {
          try {
            const errorData = JSON.parse(rawError) as { error?: string; message?: string }
            errorMessage = errorData.error || errorData.message || ""
          } catch {
            errorMessage = rawError
          }
        }

        throw new Error(errorMessage || `HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error("Create withdraw order failed:", error)
      throw error
    }
  },

  async getWithdrawMethods(): Promise<WithdrawMethod[]> {
    try {
      const response = await fetch(apiUrl("/withdraw/methods"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data: WithdrawMethodsResponse = await response.json()
      return data.methods
    } catch (error) {
      console.error("Get withdraw methods failed:", error)
      return fallbackWithdrawMethods
    }
  },

  async getVoucherConfig(): Promise<VoucherConfigResponse> {
    try {
      const response = await fetchWithAuth(apiUrl("/payments/voucher/config"), {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      })

      if (!response.ok) {
        throw new Error("Failed to fetch voucher config")
      }

      return await response.json()
    } catch (error) {
      console.error("Get voucher config failed:", error)
      return { purchase_url: "https://www.188topup.com" }
    }
  },

  async redeemVoucher(data: RedeemVoucherRequest): Promise<RedeemVoucherResponse> {
    try {
      const response = await fetchWithAuth(apiUrl("/payments/voucher/redeem"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      })

      const responseData = await response.json().catch(() => ({}))

      if (!response.ok) {
        const errorMessage = responseData.error || responseData.message || "Redeem failed"
        throw new Error(errorMessage)
      }

      return responseData
    } catch (error) {
      console.error("Redeem voucher failed:", error)
      throw error
    }
  },

  async createTokenPayOrder(data: CreateTokenPayOrderRequest): Promise<TokenPayCreateOrderResponse> {
    try {
      const response = await fetchWithAuth(apiUrl("/payments/tokenpay/order"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      })

      const responseData = await response.json().catch(() => ({}))

      if (!response.ok) {
        const errorMessage = responseData.error || responseData.message || "Create order failed"
        throw new Error(errorMessage)
      }

      return responseData
    } catch (error) {
      console.error("Create TokenPay order failed:", error)
      throw error
    }
  },
}

export interface TonOrderResponse { order_id:string; wallet_addr:string; amount_usd:number; amount_ton:number; nano_ton:number; comment:string; expires_at:string }

export interface RedeemVoucherRequest {
  amount: number
  voucher_code: string
  dst_code: string
}

export interface RedeemVoucherResponse {
  order_id: string
  amount: number
  usd_amount: number
  balance: number
  status: number
}

export interface VoucherConfigResponse {
  purchase_url: string
}

export interface TokenPayCreateOrderResponse {
  order_id: string
  pay_address: string
  pay_amount: string
  qr_code: string
  expired_at: number
  status: string
}

export interface CreateTokenPayOrderRequest {
  amount: number
  type: string
  dst_code: string
  currency: "USD" | "IDR" | "PHP"
  callback_url?: string
  chain_type: "ETH" | "TRX"
}
