package bancos

// marcas liga o nome que as pessoas veem no aplicativo aos nomes oficiais
// da lista do Banco Central. Os dois lados já estão normalizados (minúsculas,
// sem acento). "termos" é o que pode aparecer impresso no comprovante;
// "oficiais" são trechos do nome reduzido na lista.
var marcas = []struct {
	marca    string
	termos   []string
	oficiais []string
}{
	{"Nubank", []string{"nubank", "nu pagamentos", "nu financeira", "nu"}, []string{"nu pagamentos", "nu financeira", "nubank"}},
	{"Itaú", []string{"itau", "itau unibanco", "iti"}, []string{"itau"}},
	{"Bradesco", []string{"bradesco", "next"}, []string{"bradesco"}},
	{"Banco do Brasil", []string{"banco do brasil", "bco do brasil", "bb"}, []string{"bco do brasil"}},
	{"Caixa", []string{"caixa", "caixa economica", "caixa economica federal", "caixa tem"}, []string{"caixa economica federal"}},
	{"Santander", []string{"santander"}, []string{"santander"}},
	{"Inter", []string{"banco inter", "inter"}, []string{"banco inter", "inter dtvm", "inter co"}},
	{"PicPay", []string{"picpay"}, []string{"picpay"}},
	{"Mercado Pago", []string{"mercado pago", "mercadopago"}, []string{"mercado pago"}},
	{"C6 Bank", []string{"c6", "c6 bank", "banco c6"}, []string{"bco c6", "c6 ctvm"}},
	{"PagBank", []string{"pagbank", "pagseguro"}, []string{"pagseguro"}},
	{"Sicoob", []string{"sicoob", "bancoob"}, []string{"sicoob"}},
	{"Sicredi", []string{"sicredi"}, []string{"sicredi"}},
	{"Banrisul", []string{"banrisul"}, []string{"banrisul"}},
	{"BTG Pactual", []string{"btg", "btg pactual"}, []string{"btg"}},
	{"Neon", []string{"neon"}, []string{"neon"}},
	{"Stone", []string{"stone"}, []string{"stone"}},
	{"InfinitePay", []string{"infinitepay", "infinite pay", "cloudwalk"}, []string{"cloudwalk"}},
	{"Banco Pan", []string{"banco pan", "bco pan"}, []string{"banco pan", "pan finan"}},
	{"Safra", []string{"safra", "banco safra"}, []string{"safra"}},
	{"99Pay", []string{"99pay", "99 pay"}, []string{"99pay"}},
	{"RecargaPay", []string{"recargapay"}, []string{"recargapay"}},
	{"Agibank", []string{"agibank"}, []string{"agibank"}},
	{"BMG", []string{"bmg", "banco bmg"}, []string{"bco bmg"}},
	{"Banco Original", []string{"banco original"}, []string{"banco original"}},
	{"Asaas", []string{"asaas"}, []string{"asaas"}},
	{"Efí", []string{"efi", "efi bank", "gerencianet"}, []string{"efi", "gerencianet"}},
	{"Banco do Nordeste", []string{"banco do nordeste", "bnb"}, []string{"bco do nordeste"}},
	{"BRB", []string{"brb", "banco de brasilia"}, []string{"brb"}},
	{"Mercantil", []string{"mercantil", "banco mercantil"}, []string{"mercantil"}},
	{"Daycoval", []string{"daycoval"}, []string{"daycoval"}},
	{"Banestes", []string{"banestes"}, []string{"banestes"}},
	{"Will Bank", []string{"will bank", "willbank"}, []string{"will financeira", "will bank"}},
	{"Cora", []string{"cora"}, []string{"cora"}},
	{"Unicred", []string{"unicred"}, []string{"unicred"}},
	{"Cresol", []string{"cresol"}, []string{"cresol"}},
}
