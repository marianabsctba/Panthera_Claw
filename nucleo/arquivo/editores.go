package arquivo

import "strings"

// editores são programas de edição de documento ou imagem. Um comprovante
// tirado direto do aplicativo do banco não passa por nenhum deles.
var editores = []struct{ trecho, nome string }{
	{"microsoft word", "Microsoft Word"},
	{"microsoft® word", "Microsoft Word"},
	{"libreoffice", "LibreOffice"},
	{"openoffice", "OpenOffice"},
	{"wps", "WPS Office"},
	{"canva", "Canva"},
	{"photoshop", "Photoshop"},
	{"illustrator", "Illustrator"},
	{"lightroom", "Lightroom"},
	{"acrobat pro", "Acrobat"},
	{"acrobat standard", "Acrobat"},
	{"ilovepdf", "iLovePDF"},
	{"smallpdf", "Smallpdf"},
	{"sejda", "Sejda"},
	{"pdfescape", "PDFescape"},
	{"pdf candy", "PDF Candy"},
	{"sodapdf", "Soda PDF"},
	{"foxit phantompdf", "Foxit"},
	{"foxit pdf editor", "Foxit"},
	{"pdf-xchange", "PDF-XChange"},
	{"nitro", "Nitro"},
	{"pdfelement", "PDFelement"},
	{"wondershare", "Wondershare"},
	{"inkscape", "Inkscape"},
	{"gimp", "GIMP"},
	{"pixelmator", "Pixelmator"},
	{"affinity", "Affinity"},
	{"corel", "Corel"},
	{"picsart", "PicsArt"},
	{"snapseed", "Snapseed"},
	{"pixlr", "Pixlr"},
	{"photopea", "Photopea"},
	{"meitu", "Meitu"},
	{"photoroom", "PhotoRoom"},
}

// Editor devolve o nome do editor se o programa for um, ou "".
func Editor(programa string) string {
	p := strings.ToLower(programa)
	for _, e := range editores {
		if strings.Contains(p, e.trecho) {
			return e.nome
		}
	}
	return ""
}
