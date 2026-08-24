# QuantixCore (NewBayar) API 对接文档

> 文档来源：https://merchant.quantixcore.com/#/doc
> 整理日期：2026年4月15日

---

## 目录

- [QuantixCore (NewBayar) API 对接文档](#quantixcore-newbayar-api-对接文档)
  - [目录](#目录)
  - [通用说明](#通用说明)
    - [对接注意事项](#对接注意事项)
  - [接口方式及数据格式](#接口方式及数据格式)
  - [签名方式](#签名方式)
  - [代收接口](#代收接口)
    - [代收创建订单](#代收创建订单)
    - [代收查询](#代收查询)
    - [代收回调](#代收回调)
  - [代付接口](#代付接口)
    - [代付申请](#代付申请)
      - [代付说明](#代付说明)
    - [代付查询](#代付查询)
    - [代付回调](#代付回调)
  - [附录](#附录)
    - [产品编码](#产品编码)
    - [订单状态](#订单状态)
    - [币种列表](#币种列表)
    - [账户类型](#账户类型)
    - [印尼VA支持银行](#印尼va支持银行)
    - [越南网银支持银行](#越南网银支持银行)
    - [印尼代付银行列表](#印尼代付银行列表)
  - [签名示例](#签名示例)
  - [回调验签示例](#回调验签示例)

---

## 通用说明

### 对接注意事项

1. **代收订单**多次回调需要做幂等操作（防止给用户多次上分）
2. **代付订单**，网络超时问题需要处理成支付中，找渠道二次确认
3. API网关地址: https://api.quantixcore.com
---

## 接口方式及数据格式

- 使用HTTP标准的POST请求协议，请勿使用流输出等方式提交参数
- 为了接收方数据的准确性、真实性，所传输的数据均必须签名后再做上传
- 平台统一使用UTF-8编码方式
- 参数值数据格式均为JSON格式

---

## 签名方式

1. 将请求参数字段按照Ascii码方式进行升序排序（参数名a到z的顺序排序，若遇到相同的首字母，则看第二个字母，以此类推）
2. 字段值为空的不参与签名
3. 回调参数拼接时，排除sign字段
4. 集合参数按照`K1=V1&K2=V2`格式，拼接成字符串S
5. 拼接密钥信息：`S = S + '&secret=' + key`
6. 取拼接后的字符串做MD5运算，得到32位小写签名串

---

## 代收接口

### 代收创建订单

**请求地址：** `/api/open/flex/order/trade/add`

**请求参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户订单号 |
| amount | Number | 是 | 金额(单位:分) |
| code | String | 是 | 产品编码 |
| currency | String | 是 | 币种 |
| content | String | 是 | 订单内容 |
| bankCode | String | 否 | 所选银行编码（印尼VA/越南网银需要） |
| kycPayerIdNo | String | 否 | KYC付款人ID，巴西个人传CPF(纯数字)，巴西公司传CNPJ(纯数字)，非巴西不用传 |
| kycPayerName | String | 否 | KYC付款人名称，巴西个人传个人姓名，巴西公司传公司名称，非巴西不用传 |
| uid | String | 是 | fieldDesc.uid |
| clientIp | String | 是 | 用户Ip |
| callback | String | 是 | 回调地址 |
| return | String | 否 | 支付完成跳转地址 |
| sign | String | 是 | 签名 |

**响应参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| success | Boolean | 是 | true成功, false失败 |
| errorCode | String | 可选 | 错误编码 |
| message | String | 可选 | 错误信息 |
| data | Object | 可选 | 成功响应业务字段 |

**成功响应业务字段 (data)：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户订单号 |
| orderNo | String | 是 | 平台订单号 |
| payInfo | String | 是 | 支付链接 |
| raw | String | 否 | 额外支付信息，巴西为PIX二维码内容，印尼VA为VA信息，越南为收款账户 |
| amount | Number | 是 | 金额(单位:分) |
| status | String | 是 | 订单状态 |
| currency | String | 是 | 币种 |
| code | String | 是 | 产品编码 |

---

### 代收查询

**请求地址：** `/api/open/flex/order/trade/get`

**请求参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户订单号 |
| sign | String | 是 | 签名 |

**响应参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| success | Boolean | 是 | true成功, false失败 |
| errorCode | String | 可选 | 错误编码 |
| message | String | 可选 | 错误信息 |
| data | Object | 可选 | 成功响应业务字段 |

**成功响应业务字段 (data)：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户订单号 |
| orderNo | String | 是 | 平台订单号 |
| amount | Number | 是 | 金额(单位:分) |
| status | String | 是 | 订单状态 |
| currency | String | 是 | 币种 |
| code | String | 是 | 代收产品编码 |
| ref_cpf | String | 否 | 付款人CPF |
| ref_name | String | 否 | 付款人名称 |

---

### 代收回调

**请求地址：** 创建代收时传入的callback参数地址

**回调参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户订单号 |
| orderNo | String | 是 | 平台订单号 |
| amount | Number | 是 | 金额(单位:分) |
| status | String | 是 | 订单状态 |
| currency | String | 是 | 币种 |
| code | String | 是 | 代收产品编码 |
| ref_cpf | String | 否 | 付款人CPF |
| ref_name | String | 否 | 付款人名称 |
| sign | String | 是 | 签名 |

**响应要求：**
- 成功处理后，响应`success`字符串
- 未响应success时，在10s、30s、1分钟、2分钟、5分钟重试5次

---

## 代付接口

### 代付申请

**请求地址：** `/api/open/flex/order/payment/add`

**请求参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户订单号 |
| uid | String | 是 | fieldDesc.uid |
| amount | Number | 是 | 金额(单位:分) |
| currency | String | 是 | 代付币种 |
| accountType | String | 是 | 账户类型 |
| bankCode | String | 否 | 银行编码 |
| branch | String | 否 | 支行行号 |
| branchName | String | 否 | 支行名称 |
| account | String | 是 | 账号 |
| accountName | String | 是 | 账户名 |
| phone | String | 否 | 账户手机号 |
| email | String | 否 | 账户邮箱 |
| province | String | 否 | 省份 |
| city | String | 否 | 城市 |
| cpf | String | 否 | 巴西CPF/CPF_CNPJ |
| ifsc | String | 否 | 印度IFSC |
| callback | String | 否 | 回调地址 |
| sign | String | 是 | 签名 |

**响应参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| success | Boolean | 是 | true成功, false失败 |
| errorCode | String | 可选 | 错误编码 |
| message | String | 可选 | 错误信息 |
| data | Object | 可选 | 成功响应业务字段 |

**成功响应业务字段 (data)：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户单号 |
| orderNo | String | 是 | 平台代付单号 |
| amount | Number | 是 | 金额(单位:分) |
| status | String | 是 | 代付状态 |
| currency | String | 是 | 币种 |
| errorMsg | String | 否 | 失败时返回原因 |

#### 代付说明

**INRU钱包代付：**
- accountType: INRU, VIRTUAL_CURRENCY
- account: INRU钱包地址
- accountName: 玩家名称

**巴西代付：** 使用PIX相关账户类型

**印尼代付：** 使用印尼钱包或银行编码

**越南代付：** 使用越南网银或钱包

**印度代付：** 使用UPI或INRU钱包

**孟加拉代付：** 使用bKash、Nagad、Rocket钱包

**菲律宾代付：** 使用GCASH

---

### 代付查询

**请求地址：** `/api/open/flex/order/payment/get`

**请求参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户单号 |
| sign | String | 是 | 签名 |

**响应参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| success | Boolean | 是 | true成功, false失败 |
| errorCode | String | 可选 | 错误编码 |
| message | String | 可选 | 错误信息 |
| data | Object | 可选 | 成功响应业务字段 |

**成功响应业务字段 (data)：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户单号 |
| orderNo | String | 是 | 平台代付单号 |
| amount | Number | 是 | 金额(单位:分) |
| status | String | 是 | 代付状态 |
| currency | String | 是 | 币种 |
| errorMsg | String | 否 | 失败原因 |

---

### 代付回调

**请求地址：** 创建代付时传入的callback参数地址

**回调参数：**

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| merchantNo | String | 是 | 商户号 |
| merchantOrderNo | String | 是 | 商户单号 |
| orderNo | String | 是 | 平台代付单号 |
| amount | Number | 是 | 金额(单位:分) |
| status | String | 是 | 代付状态 |
| currency | String | 是 | 币种 |
| errorMsg | String | 否 | 失败时返回原因 |
| sign | String | 是 | 签名 |

**响应要求：**
- 成功处理后，响应`success`字符串
- 未响应success时，在10s、30s、1分钟、2分钟、5分钟重试5次

---

## 附录

### 产品编码

| 编码 | 描述 |
|------|------|
| PIX_QR | 巴西PIX扫码 |
| IDR_VA | 印尼VA |
| IDR_QRIS | 印尼QRIS |
| DANA | DANA |
| OVO | OVO |
| LINKAJA | LINKAJA |
| MOMO | 越南MOMO |
| ZALO | 越南ZALO |
| VN_BANK | 越南网银 |
| INR_UPI | 印度UPI |
| INRU | 印度钱包 |
| GCASH_QR | GCASH扫码 |
| GCASH_APP | GCASH唤醒 |
| GCASH_H5 | GCASH原生 |
| SHOPEEPAY | SHOPEEPAY |
| IN_NATIVE | 印度原生 |
| BDT_BKASH | 孟加拉bKash钱包 |
| BDT_NAGAD | 孟加拉Nagad钱包 |
| BDT_ROCKET | 孟加拉Rocket钱包 |
| MOMO_NATIVE | 越南MOMO原生 |
| VNQR | 越南原生QR |

---

### 订单状态

| 编码 | 描述 |
|------|------|
| WAITING_PAY | 待支付 |
| PAYING | 支付中 |
| PAID | 支付成功 |
| PAY_FAILED | 支付失败 |
| REFUND | 已退款 |

---

### 币种列表

| 编码 | 描述 |
|------|------|
| CNY | 人民币 |
| USD | 美元 |
| HKD | 港币 |
| BHD | 巴林第纳尔 |
| THB | 泰铢 |
| VND | 越南盾 |
| INR | 印度卢比 |
| NGN | 尼日利亚 |
| MXN | 墨西哥比索 |
| BRL | 雷亚尔 |
| PHP | 菲律宾比索 |
| IDR | 印尼盾 |
| COP | 哥伦比亚比索 |
| PEN | 秘鲁新索尔 |
| CLP | 智利比索 |
| PKR | 巴基斯坦卢比 |
| BDT | 孟加拉塔卡 |
| USDT | USDT |
| TRX | TRX |
| TRC10 | TRC10 |
| TRC20 | TRC20 |
| INRU | INRU |

---

### 账户类型

| 编码 | 描述 |
|------|------|
| COMPANY_BANK | 对公户 |
| PERSONAL_BANK | 个人银行卡 |
| VIRTUAL_CURRENCY | 虚拟货币 |
| UPI | 印度UPI账户 |
| PIX_EMAIL | 巴西PIX邮箱 |
| PIX_PHONE | 巴西PIX手机 |
| PIX_CPF | 巴西PIX CPF |
| PIX_CNPJ | 巴西PIX CNPJ |
| PIX_RANDOM | 巴西PIX RANDOM |
| PIX_EVP | 巴西PIX EVP |
| PIX_BANK | 巴西PIX银行 |
| GCASH | 菲律宾Gcash账户 |
| CLABE | 墨西哥Clabe账户 |
| OVO | 印尼OVO钱包 |
| DANA | 印尼DANA钱包 |
| GOPAY | 印尼GOPAY钱包 |
| SHOPEEPAY | SHOPEE钱包 |
| LINKAJA | 印尼LINKAJA钱包 |
| KASPRO | 印尼KASPRO钱包 |
| INRU | 印度INRU钱包 |
| JAZZCASH | JAZZCASH |
| EASYPAISA | EASYPAISA |
| BKASH | bKash |
| NAGAD | nagad |
| ROCKET | rocket |

---

### 印尼VA支持银行

| 银行 | 编码 |
|------|------|
| Bank BRI | BRI |
| Bank BNI | BNI |
| Bank CIMB Niaga | CIMB |
| Bank Mandiri | MANDIRI |
| Bank Neo Commerce | NEO_COM |
| Permata Bank | PERMATA |
| Bank OCBC NISP | OCBC |

---

### 越南网银支持银行

| 银行 | 编码 |
|------|------|
| NGAN HANG A CHAU | ACB |
| MARITIME BANK | MSB |
| MB BANK | MB |
| VIETCOMBANK | VCB |
| PHUONGDONG BANK | OCB |
| BIDV BANK | BIDV |
| TECHCOM BANK | TCB |

---

### 印尼代付银行列表

| 银行 | 编码 |
|------|------|
| OVO Wallet | OVO |
| DANA Wallet | DANA |
| GOPAY Wallet | GOPAY |
| GOPAY Driver | GOPAYDRIVER |
| SHOPEEPAY Wallet | SHOPEEPAY |
| LINKAJA Wallet | LINKAJA |
| KASPRO Wallet | KASPRO |
| AstraPay Wallet | ASTRAPAY |
| Bank Syariah Indonesia | BSI |
| Bank CTBC (China Trust) Indonesia | CHINATRUST |
| Bank Maybank Indocorp | BANK_MAYBANK |
| Bank Merincorp | BANK_MERIN |
| Bank Agris | BANK_AGRIS |
| Bank Yudha Bhakti | BANK_YUDHA |
| Indosat Dompetku | DOMPETKU |
| BPR KS | BPR_KS |
| Bank Harda | HARDA_INTERNASIONAL |
| Bank Mandiri Taspen Pos | MANDIRI_TASPEN |
| Bank Fama Internasional | FAMA |
| Superbank | SUPER |
| Centratama Nasional Bank | CENTRATAMA |
| Bank Index Selindo | INDEX_SELINDO |
| Bank Mayora Indonesia | MAYORA |
| Bank Multi Arta Sentosa | MULTI_ARTA_SENTOSA |
| Bank Purba Danarta | BTPN_SYARIAH |
| Bank Artos IND | ARTOS |
| Bank BCA Syariah | BCA_SYR |
| Anglomas Internasional Bank | BANK_ANGLOMAS |
| Liman International Bank | BANK_LIMAN |
| Bank Akita | BANK_AKITA |
| Bank OCBC NISP | OCBC |
| Bank Persyarikatan Indonesia | BANK_PERSY |
| Prima Master Bank | PRIMA_MASTER |
| Bank Harfa | BANK_HARFA |
| Bank Ina Perdana | BANK_INA |
| Bank Syariah Mega | MEGA_SYR |
| Bank Alfindo (Bank National Nobu) | NATIONALNOBU |
| Bank Royal Indonesia | ROYAL |
| Bank Indomonex (Bank SBI Indonesia) | BANK_INDOMONEX |
| Bank BRI Agro | AGRONIAGA |
| Bank MNC / Bank Bumiputera | BANK_BUMIPUTERA |
| Bank Bintang Manunggal | BANK_BINTANG |
| Bank Jasa Jakarta | JASA_JAKARTA |
| Bank Sri Partha | BANK_SRI_PARTHA |
| Bank Bisnis Internasional | BISNIS_INTERNASIONAL |
| Bank Syariah Mandiri(BSI) | MANDIRI_SYR |
| Bank Bukopin | BUKOPIN |
| Bank Mega | MEGA |
| Bank BJB Syariah | BJB_SYR |
| Bank Swaguna | BANK_SWAGUNA |
| JENIUS | JENIUS |
| Bank Himpunan Saudara 1906 | BANK_HIM |
| Bank Tabungan Negara (BTN) | BTN |
| Bank Tabungan Negara (BTN) UUS | BTN_UUS |
| Bank Tabungan Negara (BTN) Syariah | BTN_SYR |
| Bank QNB Kesawan (Bank QNB Indonesia) | QNB_INDONESIA |
| Bank Harmoni International | BANK_HARM |
| Halim Indonesia Bank (Bank ICBC Indonesia) | ICBC |
| Bank Windu Kentjana | CCB |
| Bank Ganesha | GANESHA |
| Bank Hagakita | BANK_HAGAKITA |
| Bank Maspion Indonesia | MASPION |
| Bank Sinarmas UUS | SINARMAS_UUS |
| Bank Metro Express (Bank Shinhan Indonesia) | SHINHAN |
| Bank Mestika Dharma | MESTIKA_DHARMA |
| Bank of India Indonesia | BANK_OF_INDIA |
| Bank Nusantara Parahyangan | NUSANTARA_PARAHYANGAN |
| Bank Sultra | BANK_SULTRA |
| Bank Sulawesi Tengah | SULAWESI |
| Bank Papua | PAPUA |
| Bank Maluku Malut | MALUKU |
| Bank NTT | BANK_NTT |
| BPD Bali | BALI |
| Bank Sulut Gorontalo | SUMSEL_DAN_BABEL_SULUT |
| BPD Sumsel Dan Babel UUS | SUMSEL_DAN_BABEL_UUS |
| Bank Kalteng | BPD_KALTENG |
| Bank Kalimantan Barat | KALIMANTAN_BARAT |
| BPD Kalimantan Selatan UUS | KALIMANTAN_SELATAN_UUS |
| BPD Kalimantan Selatan Syariah | KALIMANTAN_SELATAN_SYR |
| Bank Lampung | LAMPUNG |
| BPD Jambi | JAMBI |
| Bank Jatim | BANK_JATIM |
| Bank Jatim UUS | BANK_JATIM_UUS |
| Bank Jatim Syariah | BANK_JATIM_SYR |
| Bank Jateng | BANK_JATENG |
| BPD DIY | DAERAH_ISTIMEWA |
| Bank DKI | DKI |
| Bank DKI UUS | DKI_UUS |
| Bank DKI SYR | DKI_SYR |
| Bank Jabar dan Banten (BJB) | BANK_JABAR |
| Bank Mayapada | MAYAPADA |
| Bank JTRUST | JTRUST |
| Bank IFI | BANK_IFI |
| Bank Haga | BANK_HAGA |
| Bank Antar Daerah | BANK_ANTAR |
| Bank Ekonomi | BANK_EKONOMI |
| Bank Bumi Arta | BUMI_ARTA |
| Bank OF China | BOC |
| Bank Woori Indonesia | BANK_WOOR |
| Deutsche Bank AG. | DEUTSCHE |
| Bank ANZ Indonesia | BANK_ANZ |
| Korea Exchange Bank Danamon | BANK_DANAMON |
| Bank BNP Paribas Indonesia | BNP_PARIBAS |
| Bank Capital Indonesia | CAPITAL |
| Bank Keppel Tatlee Buana | BANK_KEPPEL |
| Bank ABN Amro | BANK_ABN |
| Standard Chartered Bank | STANDARD_CHARTERED |
| Bank Mizuho Indonesia | MIZUHO |
| Bank Resona Perdania | RESONA |
| Bank DBS Indonesia | DBS |
| Bank Sumitomo Mitsui Indonesia | MITSUI |
| The Bank of Tokyo Mitsubishi UFJ LTD | BANK_TOKYO |
| The Hongkong & Shanghai B.C. (Bank HSBC) | HSBC |
| The Hongkong & Shanghai B.C. (Bank HSBC) UUS | HSBC_UUS |
| The Hongkong & Shanghai B.C. (Bank HSBC) Syariah | HSBC_SYR |
| The Bangkok Bank Comp. LTD | BANK_COMP |
| Bank Credit Agricole Indosuez | BANK_C_AGR |
| Bank Artha Graha Internasional | ARTHA |
| ING Indonesia Bank | BANK_ING |
| Bank of America, N.A | BAML |
| JP. Morgan Chase Bank, N.A | JPMORGAN |
| Citibank | CITIBANK |
| Bank OCBC NISP | NISP |
| Bank Lippo | BANK_LIPPO |
| BANK BUANA IND | BANK_BUANA |
| Bank CIMB Niaga | CIMB |
| Bank CIMB UUS | CIMB_UUS |
| Bank CIMB Syariah | CIMB_SYR |
| Bank Panin | PANIN |
| Bank BCA | BCA |
| Permata Bank | PERMATA |
| Bank Danamon | DANAMON |
| Bank Danamon UUS | DANAMON_UUS |
| Bank Danamon SYR | DANAMON_SYR |
| Bank BNI | BNI |
| Bank Mandiri | MANDIRI |
| Bank Ekspor Indonesia | EXIMBANK |
| Bank BRI | BRI |
| Bank BNC | BNC |
| Bank Aceh Syariah | ACEH_SYR |
| BANK AMAR INDONESIA | AMAR |
| Bank Commonwealth | COMM |
| IBK Bank Indonesia | IBK |
| PT Bank KEB Hana Indonesia | HANA |
| P.T Bank Muamalat Indonesia, Tbk | MUAMALAT |
| PT Bank Oke Indonesia Tbk | OKE |
| Bank Pembangunan Daerah Banten | BANTEN |
| Bank Pembangunan Daerah Kalsel | KALSEL |
| Bank Nagari | NAGARI |
| Pencapaian Bank Sumut | SUMUT |
| BANK SULSELBAR | SULSELBAR |
| BANK VICTORIA INTERNATIONAL | VICTORIA_INTERNAL |
| BANK VICTORIA SYARIAH | VICTORIA_SYARIAH |
| BANK ALADIN SYARIAH | ALADIN_SYARIAH |
| BANK SEABANK INDONESIA | SEABANK |
| Bank Neo Commerce | NEO_COM |
| PT Bank Bengkulu | BENG_KULU |
| BANK UOB INDONESIA | UOB |
| RABOBANK INTERNASIONAL INDONESIA | RABO |
| BPD KALTIM | BPD_KAITIM |
| BPD NTB | BPD_NTB |
| BANK SWADESI | BANK_SWADESI |
| BANK TABUNGAN PENSIUNAN NASIONAL | BTPN |
| BANK NOBU | BANK_NOBU |
| BANK EKSEKUTIF | BANK_EKSEK |
| BANK BNI SYARIAH | BNI_SYR |
| BANK PANIN DUBAI SYARIAH | PANIN_SYR |
| BANK DINAR INDONESIA | DINAR_INDONESIA |
| BANK BRI SYARIAH | BRI_SYR |
| BANK SULUT | SUMSEL_DAN_BABEL |
| BPD SULSEL | BPD_SULSEL |
| BANK MITRANIAGA | BANK_MITRAN |
| BANK FINCONESIA | BANK_FINCONESIA |
| BANK PERMATA SYARIAH | PERMATA_SYR |
| BANK JAMBI SYARIAH | JAMBI_SYR |
| BPD Jambi UUS | JAMBI_UUS |
| BPD ACEH | BPD_ACEH |
| Bank Riau(BANK PEMBANGUNAN DAERAH RIAU KEPRI) | BANK_RIAU |
| BANK RIAU SYARIAH | RIAU_SYR |
| BANK KALBAR | BANK_KALBAR |
| BANK KESAWAN | BANK_KESAWAN |
| BANK SYARIAH BUKOPIN | BUKOPIN_SYR |
| Bank MNC Internasional | MNC |
| BPD Nusa Tenggara Barat | NUSA_TENGGARA_BARAT |
| Bank Tabungan Pensiunan Nasional UUS | TABUNGAN_PENSIUNAN_NASIONAL_UUS |
| BPD Nusa Tenggara Barat UUS | NUSA_TENGGARA_BARAT_UUS |
| BPD Nusa Tenggara Barat Syariah | NUSA_TENGGARA_BARAT_SYR |
| Bank OCBC NISP UUS | OCBC_UUS |
| Bank OCBC NISP Syariah | OCBC_SYR |
| Bank Sahabat Sampoerna | SAHABAT_SAMPOERNA |
| Bank Permata UUS | PERMATA_UUS |
| BPD Jawa Tengah UUS | JAWA_TENGAH_UUS |
| BPD Jawa Tengah Syariah | JAWA_TENGAH_SYR |
| Bank Maybank Syariah Indonesia | MAYBANK_SYR |
| Bank Andara | ANDARA |
| Bank Mutiara / Jtrust | MUTIARA |
| Bank Pundi Indonesia | PUNDI_INDONESIA |
| Bank Sahabat Purba Danarta | SAHABAT_PURBA_DANARTA |
| Bank Windu Kentjana Int | WINDU |
| Bank Mitra Niaga | MITRA_NIAGA |
| BPD Bengkulu | BENGKULU |
| BPD Jawa Timur UUS | JAWA_TIMUR_UUS |
| BPD Jawa Timur Syariah | JAWA_TIMUR_SYR |
| Bank Rakyat Indonesia Agroniaga (Bank Raya) | RAYA |
| BPD Riau Dan Kepri UUS | RIAU_DAN_KEPRI_UUS |
| BPD Riau Dan Kepri Syariah | RIAU_DAN_KEPRI_SYR |
| BPD Sulselbar UUS | SULSELBAR_UUS |
| BPD Sulselbar Syariah | SULSELBAR_SYR |
| Bank Pembangunan Daerah (BPD DIY) Syariah | BPD_DIY_SYR |
| Bank Rabobank International Indonesia | RABOBANK |
| Bank Arta Niaga Kencana | ARTA_NIAGA_KENCANA |
| Bank Sinar Harapan Bali / Mandiri Taspen | SINAR_HARAPAN_BALI |
| Bank Victoria Syariah | VICTORIA_SYR |
| BANK MANTAP (Mandiri Taspen) | MANTAP |
| BPD Kalimantan Barat UUS | KALIMANTAN_BARAT_UUS |
| BPD Kalimantan Barat Syariah | KALIMANTAN_BARAT_SYR |
| BPD Sumut UUS | SUMUT_UUS |
| BPD Sumut Syariah | SUMUT_SYR |
| Bank Aladin Syariah | ALADIN |
| Bank Indonesia (KPO) | BI |
| Bank Jago UUS | JAGO_UUS |
| Bank Jago Syariah | JAGO_SYR |
| BPR Eka | EKA |
| Sinarmas Syariah | SINARMAS_SYR |
| Royal Bank Scotland | SCOTLAND |
| Kustodian Sentral Efek Indonesia | KSEI |
| BPD Aceh UUS | ACEH_UUS |
| Bank Jawa Barat dan Banten Tbk | JAWA_BARAT |
| Bank SUMSEL BABEL Syariah | SUMSEL_DAN_BABEL_SYR |
| Bank Syariah Indonesia | BANK_SYAR |
| BANK DIPO INTERNATIONAL | BANK_DIPO |
| BANK TABUNGAN NEGARA (PERSERO) | TABUNGAN_NEGARA |
| THE BANGKOK BANK COMP. LTD | BANGKOK |
| UIB | BANK_UIB |
| BANK CENTRAL ASIA | BANK_ASIA |
| BANK JAGO | BANK_JAGO |
| BANK NEGARA INDONESIA | BANK_NEGARA |
| BANK OF CHINA (HONG KONG) LIMITED CABANG JAKARTA | BANK_CHIDJA |
| BANK PEMBANGUNAN DAERAH JAWA TIMUR | JAWA_TIMUR |
| BANK PEMBANGUNAN DAERAH NUSA TENGGARA TIMUR | NUSA_TENGGARA_TIMUR |
| BANK SINARMAS | SINARMAS |
| BPD JATIM UUS | JATIM_UUS |
| BPD JAWA TENGAH | JAWA_TENGAH |
| BPD KALIMANTAN TENGAH | KALIMANTAN_TENGAH |
| BPD KALIMANTAN TIMUR DAN KALIMANTAN UTARA | KALIMANTAN_TIMUR_KU |
| BPD KALIMANTAN TIMUR DAN KALIMANTAN UTARA UUS | KALIMANTAN_TIMUR_KU_UUS |
| BPD KALIMANTAN TIMUR DAN KALIMANTAN UTARA Syariah | KALIMANTAN_TIMUR_KU_SYR |
| BPD SULAWESI TENGGARA | SULAWESI_TENGGARA |
| BPD SUMATERA UTARA | SUMATERA_UTARA |
| BPD SUMATERA UTARA UUS | SUMATERA_UTARA_UUS |
| MUFG BANK | BANK_MUFG |
| PT ALLO BANK INDONESIA | ALLO |
| PT BANK BPD DIY UUS | DIY_UUS |
| PT BANK MAYBANK INDONESIA TBK | MAYBANK_TBK |
| PT BANK MAYBANK INDONESIA TBK UUS | MAYBANK_TBK_UUS |
| BANK NAGARI UUS | NAGARI_UUS |
| BANK NAGARI SYR | NAGARI_SYR |

---

## 签名示例

```python
import hashlib

def generate_sign(params, secret):
    """
    生成签名
    :param params: 请求参数字典
    :param secret: 密钥
    :return: 32位小写MD5签名
    """
    # 过滤空值
    filtered_params = {k: v for k, v in params.items() if v is not None and v != ''}
    
    # 按ASCII码升序排序
    sorted_params = sorted(filtered_params.items(), key=lambda x: x[0])
    
    # 拼接成 K1=V1&K2=V2 格式
    sign_str = '&'.join([f"{k}={v}" for k, v in sorted_params])
    
    # 拼接密钥
    sign_str += f"&secret={secret}"
    
    # MD5加密
    return hashlib.md5(sign_str.encode('utf-8')).hexdigest().lower()


# 代收创建订单示例
params = {
    "merchantNo": "M123456",
    "merchantOrderNo": "ORDER20240101001",
    "amount": 10000,
    "code": "PIX_QR",
    "currency": "BRL",
    "content": "测试订单",
    "uid": "user123",
    "clientIp": "192.168.1.1",
    "callback": "https://your-domain.com/callback"
}

secret = "your_secret_key"
sign = generate_sign(params, secret)
print(f"签名: {sign}")
```

---

## 回调验签示例

```python
def verify_callback(params, secret):
    """
    验证回调签名
    :param params: 回调参数字典（包含sign）
    :param secret: 密钥
    :return: 验签结果
    """
    received_sign = params.pop('sign', None)
    if not received_sign:
        return False
    
    # 重新生成签名
    calculated_sign = generate_sign(params, secret)
    
    return received_sign.lower() == calculated_sign.lower()


# 回调处理示例
callback_data = {
    "merchantNo": "M123456",
    "merchantOrderNo": "ORDER20240101001",
    "orderNo": "P123456789",
    "amount": 10000,
    "status": "PAID",
    "currency": "BRL",
    "code": "PIX_QR",
    "sign": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}

if verify_callback(callback_data.copy(), secret):
    print("验签成功")
    # 处理业务逻辑
    # 注意：需要做幂等处理
    return "success"
else:
    print("验签失败")
    return "fail"
```

---

*文档整理完成，如有疑问请联系平台技术支持*
