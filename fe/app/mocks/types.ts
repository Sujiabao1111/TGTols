export interface Game {
  id: string
  title: string
  category: string
  image: string
  provider: string
  minBuy: number
  maxBuy: number
  isHot?: boolean
  isNew?: boolean
  isRecommended?: boolean
}

export interface Promo {
  id: string
  title: string
  description: string
  image: string
  code: string
}

export interface ChatMessage {
  role: "user" | "model"
  text: string
  timestamp: Date
}

export enum GameCategory {
  ALL = "All",
  SLOTS = "Slots",
  LIVE = "Live Casino",
  FISHING = "Fishing",
  SPORTS = "Sports",
}

export type View =
  | "home"
  | "invite"
  | "activity"
  | "vip"
  | "wallet"
  | "profile"
  | "game"
  | "all-games"
  | "404"
  | "login"
  | "register"
  | "favorites"
  | "help-center"
  | "fairness-policy"
  | "privacy-policy"
  | "contact-us"
  | "adddesktop"
  | "bettingRank"
  | "weeklySpinWheel"
  | "dailyWeeklyChallenge"
  | "sevenDayTopup"
  | "newUserRecharge"

export type Language = "en" | "ru" | "es" | "cn" | "id" | "ph"

export interface HeroButton {
  label: string
  action: string
  primary?: boolean
}

export interface HeroSlide {
  id: number
  title: string
  subtitle?: string
  tag?: string
  image: string
  color: string
  jumpLink?: string
  buttons?: HeroButton[]
}

export type WalletTab = "deposit" | "withdraw" | "records"

export interface Provider {
  id: number
  platform_code?: string
  code: string
  name: string
  status: number
  type_id: number
  created_at: string
}

export interface GameType {
  id: number
  name: string
  sort: number
  code: string
  status: number
}

export interface GameOptionsResponse {
  code: number
  data: {
    providers: Provider[]
    game_types: GameType[]
  }
  message: string
}

export interface ServerGame {
  id: number | string
  platform_code?: string
  provider_id?: number
  game_type_id?: number
  game_code: string
  name: {
    CN: string
    EN: string
    KR?: string
    TH?: string
    ZH: string
    RU?: string
  }
  img_url: string
  views?: number
  sort?: number
  status?: number
  created_at?: string
  provider: Provider
  game_type?: string
  is_lobby?: boolean
  is_favorite?: boolean
}

export interface GameListResponse {
  code: number
  data: {
    list: ServerGame[]
    page: number
    page_size: number
    total: number
    total_pages: number
  }
  message: string
}

export interface GameLobbyResponse {
  code: number
  data: {
    url: string
    html?: string
    launch_content_type?: "url" | "html"
    platform_code?: string
    local_balance?: number
    game_balance?: number
    desktop_reward_claim?: ActivityClaimRecordResponse
    desktop_reward_checked?: boolean
  }
  message: string
}

export interface GameLaunchResponse {
  code: number
  data: {
    url: string
    html?: string
    launch_content_type?: "url" | "html"
    platform_code?: string
    local_balance?: number
    game_balance?: number
    is_favorite?: boolean
    desktop_reward_claim?: ActivityClaimRecordResponse
    desktop_reward_checked?: boolean
  }
  message: string
}

export interface ActivityClaimRecordResponse {
  success: boolean
  activity_type: string
  claimed: boolean
  already_claimed: boolean
  claimed_at: string
  reward_amount: number
  balance: number
  reward_type?: string
  coupon_id?: number
  message: string
}

export interface AddDesktopInsuranceSettlementResponse {
  triggered: boolean
  entry_amount_u: number
  exit_amount_u: number
  loss_amount_u: number
  compensation_amount_u: number
  current_balance: number
  coupon_code?: string
  message?: string
}

export interface RecycleBalanceResponse {
  code: number
  data: {
    recycled_amount: number
    current_balance: number
    insurance?: AddDesktopInsuranceSettlementResponse | null
  }
  message: string
}

export interface UserInfoResponse {
  code: number
  data: {
    uid: number
    username: string
    invite_code: string
    balance: number
    vip_level: number
    level: number
    total_deposit?: number
    remaining_wager?: number
    withdrawable_balance?: number
  }
  message: string
}

export interface FavoriteGamesResponse {
  code: number
  data: {
    list: ServerGame[]
    total: number
    page: number
  }
  message: string
}

export interface DoFavoriteResponse {
  code: number
  data?: {
    is_favorite: boolean
  }
  message: string
}

export interface InviteInfoResponse {
  code: number
  data: {
    invite_code: string
    share_link: string
    direct_count: number
    team_count: number
    total_earn: number
  }
  message: string
}

export interface BannerResponse {
  code: number
  data: HeroSlide[]
  message: string
}

export type SortType = "POPULAR" | "RECOMMEND" | "NEW"

export interface SlotGamesListResponse {
  code: number
  message: string
  data: {
    list: ServerGame[]
    total: number
    page: number
    page_size: number
    total_pages: number
    sort_type: SortType
  }
}
