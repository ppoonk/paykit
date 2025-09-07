package paykit

type Currency string

// 零小数货币，货币的最小单位 = 首选货币单位。例如 JPY 标准单位和最小单位都是日元
var ZeroDecimalCurrency = []Currency{
	CurrencyBIF,
	CurrencyCLP,
	CurrencyDJF,
	CurrencyGNF,
	CurrencyJPY,
	CurrencyKMF,
	CurrencyKRW,
	CurrencyMGA,
	CurrencyPYG,
	CurrencyRWF,
	CurrencyUGX,
	CurrencyVND,
	CurrencyVUV,
	CurrencyXAF,
	CurrencyXOF,
	CurrencyXPF,
}

// ISO 4217 三个字母的货币代码
const (
	CurrencyAED Currency = "AED" // United Arab Emirates Dirham
	CurrencyAFN Currency = "AFN" // Afghan Afghani
	CurrencyALL Currency = "ALL" // Albanian Lek
	CurrencyAMD Currency = "AMD" // Armenian Dram
	CurrencyANG Currency = "ANG" // Netherlands Antillean Gulden
	CurrencyAOA Currency = "AOA" // Angolan Kwanza
	CurrencyARS Currency = "ARS" // Argentine Peso
	CurrencyAUD Currency = "AUD" // Australian Dollar
	CurrencyAWG Currency = "AWG" // Aruban Florin
	CurrencyAZN Currency = "AZN" // Azerbaijani Manat
	CurrencyBAM Currency = "BAM" // Bosnia & Herzegovina Convertible Mark
	CurrencyBBD Currency = "BBD" // Barbadian Dollar
	CurrencyBDT Currency = "BDT" // Bangladeshi Taka
	CurrencyBGN Currency = "BGN" // Bulgarian Lev
	CurrencyBIF Currency = "BIF" // Burundian Franc
	CurrencyBMD Currency = "BMD" // Bermudian Dollar
	CurrencyBND Currency = "BND" // Brunei Dollar
	CurrencyBOB Currency = "BOB" // Bolivian Boliviano
	CurrencyBRL Currency = "BRL" // Brazilian Real
	CurrencyBSD Currency = "BSD" // Bahamian Dollar
	CurrencyBWP Currency = "BWP" // Botswana Pula
	CurrencyBZD Currency = "BZD" // Belize Dollar
	CurrencyCAD Currency = "CAD" // Canadian Dollar
	CurrencyCDF Currency = "CDF" // Congolese Franc
	CurrencyCHF Currency = "CHF" // Swiss Franc
	CurrencyCLP Currency = "CLP" // Chilean Peso
	CurrencyCNY Currency = "CNY" // Chinese Renminbi Yuan
	CurrencyCOP Currency = "COP" // Colombian Peso
	CurrencyCRC Currency = "CRC" // Costa Rican Colón
	CurrencyCVE Currency = "CVE" // Cape Verdean Escudo
	CurrencyCZK Currency = "CZK" // Czech Koruna
	CurrencyDJF Currency = "DJF" // Djiboutian Franc
	CurrencyDKK Currency = "DKK" // Danish Krone
	CurrencyDOP Currency = "DOP" // Dominican Peso
	CurrencyDZD Currency = "DZD" // Algerian Dinar
	CurrencyEEK Currency = "EEK" // Estonian Kroon
	CurrencyEGP Currency = "EGP" // Egyptian Pound
	CurrencyETB Currency = "ETB" // Ethiopian Birr
	CurrencyEUR Currency = "EUR" // Euro
	CurrencyFJD Currency = "FJD" // Fijian Dollar
	CurrencyFKP Currency = "FKP" // Falkland Islands Pound
	CurrencyGBP Currency = "GBP" // British Pound
	CurrencyGEL Currency = "GEL" // Georgian Lari
	CurrencyGIP Currency = "GIP" // Gibraltar Pound
	CurrencyGMD Currency = "GMD" // Gambian Dalasi
	CurrencyGNF Currency = "GNF" // Guinean Franc
	CurrencyGTQ Currency = "GTQ" // Guatemalan Quetzal
	CurrencyGYD Currency = "GYD" // Guyanese Dollar
	CurrencyHKD Currency = "HKD" // Hong Kong Dollar
	CurrencyHNL Currency = "HNL" // Honduran Lempira
	CurrencyHRK Currency = "HRK" // Croatian Kuna
	CurrencyHTG Currency = "HTG" // Haitian Gourde
	CurrencyHUF Currency = "HUF" // Hungarian Forint
	CurrencyIDR Currency = "IDR" // Indonesian Rupiah
	CurrencyILS Currency = "ILS" // Israeli New Sheqel
	CurrencyINR Currency = "INR" // Indian Rupee
	CurrencyISK Currency = "ISK" // Icelandic Króna
	CurrencyJMD Currency = "JMD" // Jamaican Dollar
	CurrencyJPY Currency = "JPY" // Japanese Yen
	CurrencyKES Currency = "KES" // Kenyan Shilling
	CurrencyKGS Currency = "KGS" // Kyrgyzstani Som
	CurrencyKHR Currency = "KHR" // Cambodian Riel
	CurrencyKMF Currency = "KMF" // Comorian Franc
	CurrencyKRW Currency = "KRW" // South Korean Won
	CurrencyKYD Currency = "KYD" // Cayman Islands Dollar
	CurrencyKZT Currency = "KZT" // Kazakhstani Tenge
	CurrencyLAK Currency = "LAK" // Lao Kip
	CurrencyLBP Currency = "LBP" // Lebanese Pound
	CurrencyLKR Currency = "LKR" // Sri Lankan Rupee
	CurrencyLRD Currency = "LRD" // Liberian Dollar
	CurrencyLSL Currency = "LSL" // Lesotho Loti
	CurrencyLTL Currency = "LTL" // Lithuanian Litas
	CurrencyLVL Currency = "LVL" // Latvian Lats
	CurrencyMAD Currency = "MAD" // Moroccan Dirham
	CurrencyMDL Currency = "MDL" // Moldovan Leu
	CurrencyMGA Currency = "MGA" // Malagasy Ariary
	CurrencyMKD Currency = "MKD" // Macedonian Denar
	CurrencyMNT Currency = "MNT" // Mongolian Tögrög
	CurrencyMOP Currency = "MOP" // Macanese Pataca
	CurrencyMRO Currency = "MRO" // Mauritanian Ouguiya
	CurrencyMUR Currency = "MUR" // Mauritian Rupee
	CurrencyMVR Currency = "MVR" // Maldivian Rufiyaa
	CurrencyMWK Currency = "MWK" // Malawian Kwacha
	CurrencyMXN Currency = "MXN" // Mexican Peso
	CurrencyMYR Currency = "MYR" // Malaysian Ringgit
	CurrencyMZN Currency = "MZN" // Mozambican Metical
	CurrencyNAD Currency = "NAD" // Namibian Dollar
	CurrencyNGN Currency = "NGN" // Nigerian Naira
	CurrencyNIO Currency = "NIO" // Nicaraguan Córdoba
	CurrencyNOK Currency = "NOK" // Norwegian Krone
	CurrencyNPR Currency = "NPR" // Nepalese Rupee
	CurrencyNZD Currency = "NZD" // New Zealand Dollar
	CurrencyPAB Currency = "PAB" // Panamanian Balboa
	CurrencyPEN Currency = "PEN" // Peruvian Nuevo Sol
	CurrencyPGK Currency = "PGK" // Papua New Guinean Kina
	CurrencyPHP Currency = "PHP" // Philippine Peso
	CurrencyPKR Currency = "PKR" // Pakistani Rupee
	CurrencyPLN Currency = "PLN" // Polish Złoty
	CurrencyPYG Currency = "PYG" // Paraguayan Guaraní
	CurrencyQAR Currency = "QAR" // Qatari Riyal
	CurrencyRON Currency = "RON" // Romanian Leu
	CurrencyRSD Currency = "RSD" // Serbian Dinar
	CurrencyRUB Currency = "RUB" // Russian Ruble
	CurrencyRWF Currency = "RWF" // Rwandan Franc
	CurrencySAR Currency = "SAR" // Saudi Riyal
	CurrencySBD Currency = "SBD" // Solomon Islands Dollar
	CurrencySCR Currency = "SCR" // Seychellois Rupee
	CurrencySEK Currency = "SEK" // Swedish Krona
	CurrencySGD Currency = "SGD" // Singapore Dollar
	CurrencySHP Currency = "SHP" // Saint Helenian Pound
	CurrencySLL Currency = "SLL" // Sierra Leonean Leone
	CurrencySOS Currency = "SOS" // Somali Shilling
	CurrencySRD Currency = "SRD" // Surinamese Dollar
	CurrencySTN Currency = "STN" // São Tomé and Príncipe Dobra
	CurrencySVC Currency = "SVC" // Salvadoran Colón
	CurrencySZL Currency = "SZL" // Swazi Lilangeni
	CurrencyTHB Currency = "THB" // Thai Baht
	CurrencyTJS Currency = "TJS" // Tajikistani Somoni
	CurrencyTOP Currency = "TOP" // Tongan Paʻanga
	CurrencyTRY Currency = "TRY" // Turkish Lira
	CurrencyTTD Currency = "TTD" // Trinidad and Tobago Dollar
	CurrencyTWD Currency = "TWD" // New Taiwan Dollar
	CurrencyTZS Currency = "TZS" // Tanzanian Shilling
	CurrencyUAH Currency = "UAH" // Ukrainian Hryvnia
	CurrencyUGX Currency = "UGX" // Ugandan Shilling
	CurrencyUSD Currency = "USD" // United States Dollar
	CurrencyUYU Currency = "UYU" // Uruguayan Peso
	CurrencyUZS Currency = "UZS" // Uzbekistani Som
	CurrencyVES Currency = "VES" // Venezuelan Bolívar
	CurrencyVND Currency = "VND" // Vietnamese Đồng
	CurrencyVUV Currency = "VUV" // Vanuatu Vatu
	CurrencyWST Currency = "WST" // Samoan Tala
	CurrencyXAF Currency = "XAF" // Central African Cfa Franc
	CurrencyXCD Currency = "XCD" // East Caribbean Dollar
	CurrencyXOF Currency = "XOF" // West African Cfa Franc
	CurrencyXPF Currency = "XPF" // Cfp Franc
	CurrencyYER Currency = "YER" // Yemeni Rial
	CurrencyZAR Currency = "ZAR" // South African Rand
	CurrencyZMW Currency = "ZMW" // Zambian Kwacha
)
