package main

// Metadatos de líderes y gimnasios que no vienen en las fuentes de equipos.

// Identificadores que se repiten en varias tablas de este fichero.
const (
	leaderLtSurge     = "lt-surge"
	leaderCrasherWake = "crasher-wake"

	gymPewter    = "pewter-gym"
	gymCerulean  = "cerulean-gym"
	gymVermilion = "vermilion-gym"
	gymCeladon   = "celadon-gym"
	gymFuchsia   = "fuchsia-gym"
	gymSaffron   = "saffron-gym"
	gymCinnabar  = "cinnabar-gym"
	gymViridian  = "viridian-gym"

	gymOreburgh  = "oreburgh-gym"
	gymEterna    = "eterna-gym"
	gymVeilstone = "veilstone-gym"
	gymPastoria  = "pastoria-gym"
	gymHearthome = "hearthome-gym"
	gymCanalave  = "canalave-gym"
	gymSnowpoint = "snowpoint-gym"
	gymSunyshore = "sunyshore-gym"

	gymStriaton   = "striaton-gym"
	gymCastelia   = "castelia-gym"
	gymNimbasa    = "nimbasa-gym"
	gymDriftveil  = "driftveil-gym"
	gymMistralton = "mistralton-gym"
	gymOpelucid   = "opelucid-gym"

	gymStowOnSide = "stow-on-side-stadium"
	gymCirchester = "circhester-stadium"

	notePostGame  = "Kanto (post-game)"
	noteFreeOrder = "Orden libre; orden por nivel"
)

type leaderDef struct{ id, en, es, gender string }

var leaderDefs = []leaderDef{
	// Kanto
	{"brock", "Brock", "Brock", "M"},
	{"misty", "Misty", "Misty", "F"},
	{leaderLtSurge, "Lt. Surge", "Lt. Surge", "M"},
	{"erika", "Erika", "Erika", "F"},
	{"koga", "Koga", "Koga", "M"},
	{"sabrina", "Sabrina", "Sabrina", "F"},
	{"blaine", "Blaine", "Blaine", "M"},
	{"giovanni", "Giovanni", "Giovanni", "M"},
	{"janine", "Janine", "Sachiko", "F"},
	{"blue", "Blue", "Azul", "M"},
	// Johto
	{"falkner", "Falkner", "Pegaso", "M"},
	{"bugsy", "Bugsy", "Antón", "M"},
	{"whitney", "Whitney", "Blanca", "F"},
	{"morty", "Morty", "Morti", "M"},
	{"chuck", "Chuck", "Aníbal", "M"},
	{"jasmine", "Jasmine", "Yasmina", "F"},
	{"pryce", "Pryce", "Fredo", "M"},
	{"clair", "Clair", "Débora", "F"},
	// Hoenn
	{"roxanne", "Roxanne", "Petra", "F"},
	{"brawly", "Brawly", "Marcial", "M"},
	{"wattson", "Wattson", "Erico", "M"},
	{"flannery", "Flannery", "Candela", "F"},
	{"norman", "Norman", "Norman", "M"},
	{"winona", "Winona", "Alana", "F"},
	{"tate-and-liza", "Tate & Liza", "Vito y Leti", "MF"},
	{"wallace", "Wallace", "Plubio", "M"},
	{"juan", "Juan", "Galano", "M"},
	// Sinnoh
	{"roark", "Roark", "Roco", "M"},
	{"gardenia", "Gardenia", "Gardenia", "F"},
	{"maylene", "Maylene", "Brega", "F"},
	{leaderCrasherWake, "Crasher Wake", "Mananti", "M"},
	{"fantina", "Fantina", "Fantina", "F"},
	{"byron", "Byron", "Acerón", "M"},
	{"candice", "Candice", "Inverna", "F"},
	{"volkner", "Volkner", "Lectro", "M"},
	// Teselia
	{"cilan", "Cilan", "Millo", "M"},
	{"chili", "Chili", "Zeo", "M"},
	{"cress", "Cress", "Maiz", "M"},
	{"lenora", "Lenora", "Aloe", "F"},
	{"burgh", "Burgh", "Camus", "M"},
	{"elesa", "Elesa", "Camila", "F"},
	{"clay", "Clay", "Yakón", "M"},
	{"skyla", "Skyla", "Gerania", "F"},
	{"brycen", "Brycen", "Junco", "M"},
	{"drayden", "Drayden", "Lirio", "M"},
	{"iris", "Iris", "Iris", "F"},
	{"cheren", "Cheren", "Cheren", "M"},
	{"roxie", "Roxie", "Hiedra", "F"},
	{"marlon", "Marlon", "Ciprián", "M"},
	// Kalos
	{"viola", "Viola", "Violeta", "F"},
	{"grant", "Grant", "Lino", "M"},
	{"korrina", "Korrina", "Corelia", "F"},
	{"ramos", "Ramos", "Amaro", "M"},
	{"clemont", "Clemont", "Lem", "M"},
	{"valerie", "Valerie", "Valeria", "F"},
	{"olympia", "Olympia", "Ástrid", "F"},
	{"wulfric", "Wulfric", "Édel", "M"},
	// Galar
	{"milo", "Milo", "Percy", "M"},
	{"nessa", "Nessa", "Cathy", "F"},
	{"kabu", "Kabu", "Naboru", "M"},
	{"bea", "Bea", "Judith", "F"},
	{"allister", "Allister", "Alistair", "M"},
	{"opal", "Opal", "Sally", "F"},
	{"gordie", "Gordie", "Morris", "M"},
	{"melony", "Melony", "Mel", "F"},
	{"piers", "Piers", "Nerio", "M"},
	{"raihan", "Raihan", "Roy", "M"},
	// Paldea
	{"katy", "Katy", "Araceli", "F"},
	{"brassius", "Brassius", "Brais", "M"},
	{"iono", "Iono", "e-Nigma", "F"},
	{"kofu", "Kofu", "Fuco", "M"},
	{"larry", "Larry", "Laureano", "M"},
	{"ryme", "Ryme", "Lima", "F"},
	{"tulip", "Tulip", "Tuli", "F"},
	{"grusha", "Grusha", "Grusha", "M"},
}

// Gimnasio: identificador, ciudad (inglés), identificador de location en PokeAPI, región, medalla por defecto.
type gymDef struct{ id, city, location, region, badge string }

var gymDefs = []gymDef{
	{gymPewter, "Pewter City", "pewter-city", "kanto", "Boulder Badge"},
	{gymCerulean, "Cerulean City", "cerulean-city", "kanto", "Cascade Badge"},
	{gymVermilion, "Vermilion City", "vermilion-city", "kanto", "Thunder Badge"},
	{gymCeladon, "Celadon City", "celadon-city", "kanto", "Rainbow Badge"},
	{gymFuchsia, "Fuchsia City", "fuchsia-city", "kanto", "Soul Badge"},
	{gymSaffron, "Saffron City", "saffron-city", "kanto", "Marsh Badge"},
	{gymCinnabar, "Cinnabar Island", "cinnabar-island", "kanto", "Volcano Badge"},
	{gymViridian, "Viridian City", "viridian-city", "kanto", "Earth Badge"},

	{"violet-gym", "Violet City", "violet-city", "johto", "Zephyr Badge"},
	{"azalea-gym", "Azalea Town", "azalea-town", "johto", "Hive Badge"},
	{"goldenrod-gym", "Goldenrod City", "goldenrod-city", "johto", "Plain Badge"},
	{"ecruteak-gym", "Ecruteak City", "ecruteak-city", "johto", "Fog Badge"},
	{"cianwood-gym", "Cianwood City", "cianwood-city", "johto", "Storm Badge"},
	{"olivine-gym", "Olivine City", "olivine-city", "johto", "Mineral Badge"},
	{"mahogany-gym", "Mahogany Town", "mahogany-town", "johto", "Glacier Badge"},
	{"blackthorn-gym", "Blackthorn City", "blackthorn-city", "johto", "Rising Badge"},

	{"rustboro-gym", "Rustboro City", "rustboro-city", "hoenn", "Stone Badge"},
	{"dewford-gym", "Dewford Town", "dewford-town", "hoenn", "Knuckle Badge"},
	{"mauville-gym", "Mauville City", "mauville-city", "hoenn", "Dynamo Badge"},
	{"lavaridge-gym", "Lavaridge Town", "lavaridge-town", "hoenn", "Heat Badge"},
	{"petalburg-gym", "Petalburg City", "petalburg-city", "hoenn", "Balance Badge"},
	{"fortree-gym", "Fortree City", "fortree-city", "hoenn", "Feather Badge"},
	{"mossdeep-gym", "Mossdeep City", "mossdeep-city", "hoenn", "Mind Badge"},
	{"sootopolis-gym", "Sootopolis City", "sootopolis-city", "hoenn", "Rain Badge"},

	{gymOreburgh, "Oreburgh City", "oreburgh-city", "sinnoh", "Coal Badge"},
	{gymEterna, "Eterna City", "eterna-city", "sinnoh", "Forest Badge"},
	{gymVeilstone, "Veilstone City", "veilstone-city", "sinnoh", "Cobble Badge"},
	{gymPastoria, "Pastoria City", "pastoria-city", "sinnoh", "Fen Badge"},
	{gymHearthome, "Hearthome City", "hearthome-city", "sinnoh", "Relic Badge"},
	{gymCanalave, "Canalave City", "canalave-city", "sinnoh", "Mine Badge"},
	{gymSnowpoint, "Snowpoint City", "snowpoint-city", "sinnoh", "Icicle Badge"},
	{gymSunyshore, "Sunyshore City", "sunyshore-city", "sinnoh", "Beacon Badge"},

	{gymStriaton, "Striaton City", "striaton-city", "unova", "Trio Badge"},
	{"nacrene-gym", "Nacrene City", "nacrene-city", "unova", "Basic Badge"},
	{gymCastelia, "Castelia City", "castelia-city", "unova", "Insect Badge"},
	{gymNimbasa, "Nimbasa City", "nimbasa-city", "unova", "Bolt Badge"},
	{gymDriftveil, "Driftveil City", "driftveil-city", "unova", "Quake Badge"},
	{gymMistralton, "Mistralton City", "mistralton-city", "unova", "Jet Badge"},
	{"icirrus-gym", "Icirrus City", "icirrus-city", "unova", "Freeze Badge"},
	{gymOpelucid, "Opelucid City", "opelucid-city", "unova", "Legend Badge"},
	{"aspertia-gym", "Aspertia City", "aspertia-city", "unova", "Basic Badge"},
	{"virbank-gym", "Virbank City", "virbank-city", "unova", "Toxic Badge"},
	{"humilau-gym", "Humilau City", "humilau-city", "unova", "Wave Badge"},

	{"santalune-gym", "Santalune City", "santalune-city", "kalos", "Bug Badge"},
	{"cyllage-gym", "Cyllage City", "cyllage-city", "kalos", "Cliff Badge"},
	{"shalour-gym", "Shalour City", "shalour-city", "kalos", "Rumble Badge"},
	{"coumarine-gym", "Coumarine City", "coumarine-city", "kalos", "Plant Badge"},
	{"lumiose-gym", "Lumiose City", "lumiose-city", "kalos", "Voltage Badge"},
	{"laverre-gym", "Laverre City", "laverre-city", "kalos", "Fairy Badge"},
	{"anistar-gym", "Anistar City", "anistar-city", "kalos", "Psychic Badge"},
	{"snowbelle-gym", "Snowbelle City", "snowbelle-city", "kalos", "Iceberg Badge"},

	{"turffield-stadium", "Turffield", "turffield", "galar", "Grass Badge"},
	{"hulbury-stadium", "Hulbury", "hulbury", "galar", "Water Badge"},
	{"motostoke-stadium", "Motostoke", "motostoke", "galar", "Fire Badge"},
	{gymStowOnSide, "Stow-on-Side", "stow-on-side", "galar", "Fighting Badge"},
	{"ballonlea-stadium", "Ballonlea", "ballonlea", "galar", "Fairy Badge"},
	{gymCirchester, "Circhester", "circhester", "galar", "Rock Badge"},
	{"spikemuth-gym", "Spikemuth", "spikemuth", "galar", "Dark Badge"},
	{"hammerlocke-stadium", "Hammerlocke", "hammerlocke", "galar", "Dragon Badge"},

	{"cortondo-gym", "Cortondo", "cortondo", "paldea", "Bug Badge"},
	{"artazon-gym", "Artazon", "artazon", "paldea", "Grass Badge"},
	{"levincia-gym", "Levincia", "levincia", "paldea", "Electric Badge"},
	{"cascarrafa-gym", "Cascarrafa", "cascarrafa", "paldea", "Water Badge"},
	{"medali-gym", "Medali", "medali", "paldea", "Normal Badge"},
	{"montenevera-gym", "Montenevera", "montenevera", "paldea", "Ghost Badge"},
	{"alfornada-gym", "Alfornada", "alfornada", "paldea", "Psychic Badge"},
	{"glaseado-gym", "Glaseado", "", "paldea", "Ice Badge"},
}

// Una plaza de líder en un juego concreto.
type slotDef struct {
	leader, gym string
	order       int
	typ         string // identificador del tipo; "" = sin tipo
	version     string // "" = todas las versiones del grupo
	badge       string // "" = la del gimnasio
	notes       string
}

var kantoSlots = []slotDef{
	{"brock", gymPewter, 1, "rock", "", "", ""},
	{"misty", gymCerulean, 2, "water", "", "", ""},
	{leaderLtSurge, gymVermilion, 3, "electric", "", "", ""},
	{"erika", gymCeladon, 4, "grass", "", "", ""},
	{"koga", gymFuchsia, 5, "poison", "", "", ""},
	{"sabrina", gymSaffron, 6, "psychic", "", "", ""},
	{"blaine", gymCinnabar, 7, "fire", "", "", ""},
	{"giovanni", gymViridian, 8, "ground", "", "", ""},
}

var johtoSlots = []slotDef{
	{"falkner", "violet-gym", 1, "flying", "", "", ""},
	{"bugsy", "azalea-gym", 2, "bug", "", "", ""},
	{"whitney", "goldenrod-gym", 3, "normal", "", "", ""},
	{"morty", "ecruteak-gym", 4, "ghost", "", "", ""},
	{"chuck", "cianwood-gym", 5, "fighting", "", "", ""},
	{"jasmine", "olivine-gym", 6, "steel", "", "", ""},
	{"pryce", "mahogany-gym", 7, "ice", "", "", ""},
	{"clair", "blackthorn-gym", 8, "dragon", "", "", ""},
	// Kanto en GSC/HGSS: medallas 9-16
	{"brock", gymPewter, 9, "rock", "", "", notePostGame},
	{"misty", gymCerulean, 10, "water", "", "", notePostGame},
	{leaderLtSurge, gymVermilion, 11, "electric", "", "", notePostGame},
	{"erika", gymCeladon, 12, "grass", "", "", notePostGame},
	{"janine", gymFuchsia, 13, "poison", "", "", notePostGame},
	{"sabrina", gymSaffron, 14, "psychic", "", "", notePostGame},
	{"blaine", gymCinnabar, 15, "fire", "", "", notePostGame + "; el gimnasio está en las Islas Espuma"},
	{"blue", gymViridian, 16, "", "", "", notePostGame},
}

func hoennSlots(champ string) []slotDef {
	return []slotDef{
		{"roxanne", "rustboro-gym", 1, "rock", "", "", ""},
		{"brawly", "dewford-gym", 2, "fighting", "", "", ""},
		{"wattson", "mauville-gym", 3, "electric", "", "", ""},
		{"flannery", "lavaridge-gym", 4, "fire", "", "", ""},
		{"norman", "petalburg-gym", 5, "normal", "", "", ""},
		{"winona", "fortree-gym", 6, "flying", "", "", ""},
		{"tate-and-liza", "mossdeep-gym", 7, "psychic", "", "", ""},
		{champ, "sootopolis-gym", 8, "water", "", "", ""},
	}
}

var dpSlots = []slotDef{
	{"roark", gymOreburgh, 1, "rock", "", "", ""},
	{"gardenia", gymEterna, 2, "grass", "", "", ""},
	{"maylene", gymVeilstone, 3, "fighting", "", "", ""},
	{leaderCrasherWake, gymPastoria, 4, "water", "", "", ""},
	{"fantina", gymHearthome, 5, "ghost", "", "", ""},
	{"byron", gymCanalave, 6, "steel", "", "", ""},
	{"candice", gymSnowpoint, 7, "ice", "", "", ""},
	{"volkner", gymSunyshore, 8, "electric", "", "", ""},
}

var ptSlots = []slotDef{
	{"roark", gymOreburgh, 1, "rock", "", "", ""},
	{"gardenia", gymEterna, 2, "grass", "", "", ""},
	{"fantina", gymHearthome, 3, "ghost", "", "", ""},
	{"maylene", gymVeilstone, 4, "fighting", "", "", ""},
	{leaderCrasherWake, gymPastoria, 5, "water", "", "", ""},
	{"byron", gymCanalave, 6, "steel", "", "", ""},
	{"candice", gymSnowpoint, 7, "ice", "", "", ""},
	{"volkner", gymSunyshore, 8, "electric", "", "", ""},
}

// gameSlots: version group -> plazas de líder
var gameSlots = map[string][]slotDef{
	"red-blue":                        kantoSlots,
	"yellow":                          kantoSlots,
	"gold-silver":                     johtoSlots,
	"crystal":                         johtoSlots,
	"ruby-sapphire":                   hoennSlots("wallace"),
	"emerald":                         hoennSlots("juan"),
	"firered-leafgreen":               kantoSlots,
	"diamond-pearl":                   dpSlots,
	"platinum":                        ptSlots,
	"heartgold-soulsilver":            johtoSlots,
	"lets-go-pikachu-lets-go-eevee":   kantoSlots,
	"omega-ruby-alpha-sapphire":       hoennSlots("wallace"),
	"brilliant-diamond-shining-pearl": dpSlots,
	"black-white": {
		{"chili", gymStriaton, 1, "fire", "", "", "Si el jugador eligió a Snivy"},
		{"cress", gymStriaton, 1, "water", "", "", "Si el jugador eligió a Tepig"},
		{"cilan", gymStriaton, 1, "grass", "", "", "Si el jugador eligió a Oshawott"},
		{"lenora", "nacrene-gym", 2, "normal", "", "", ""},
		{"burgh", gymCastelia, 3, "bug", "", "", ""},
		{"elesa", gymNimbasa, 4, "electric", "", "", ""},
		{"clay", gymDriftveil, 5, "ground", "", "", ""},
		{"skyla", gymMistralton, 6, "flying", "", "", ""},
		{"brycen", "icirrus-gym", 7, "ice", "", "", ""},
		{"drayden", gymOpelucid, 8, "dragon", "black", "", ""},
		{"iris", gymOpelucid, 8, "dragon", "white", "", ""},
	},
	"black-2-white-2": {
		{"cheren", "aspertia-gym", 1, "normal", "", "", ""},
		{"roxie", "virbank-gym", 2, "poison", "", "", ""},
		{"burgh", gymCastelia, 3, "bug", "", "", ""},
		{"elesa", gymNimbasa, 4, "electric", "", "", ""},
		{"clay", gymDriftveil, 5, "ground", "", "", ""},
		{"skyla", gymMistralton, 6, "flying", "", "", ""},
		{"drayden", gymOpelucid, 7, "dragon", "", "", ""},
		{"marlon", "humilau-gym", 8, "water", "", "", ""},
	},
	"x-y": {
		{"viola", "santalune-gym", 1, "bug", "", "", ""},
		{"grant", "cyllage-gym", 2, "rock", "", "", ""},
		{"korrina", "shalour-gym", 3, "fighting", "", "", ""},
		{"ramos", "coumarine-gym", 4, "grass", "", "", ""},
		{"clemont", "lumiose-gym", 5, "electric", "", "", ""},
		{"valerie", "laverre-gym", 6, "fairy", "", "", ""},
		{"olympia", "anistar-gym", 7, "psychic", "", "", ""},
		{"wulfric", "snowbelle-gym", 8, "ice", "", "", ""},
	},
	"sword-shield": {
		{"milo", "turffield-stadium", 1, "grass", "", "", ""},
		{"nessa", "hulbury-stadium", 2, "water", "", "", ""},
		{"kabu", "motostoke-stadium", 3, "fire", "", "", ""},
		{"bea", gymStowOnSide, 4, "fighting", "sword", "Fighting Badge", ""},
		{"allister", gymStowOnSide, 4, "ghost", "shield", "Ghost Badge", ""},
		{"opal", "ballonlea-stadium", 5, "fairy", "", "", ""},
		{"gordie", gymCirchester, 6, "rock", "sword", "Rock Badge", ""},
		{"melony", gymCirchester, 6, "ice", "shield", "Ice Badge", ""},
		{"piers", "spikemuth-gym", 7, "dark", "", "", ""},
		{"raihan", "hammerlocke-stadium", 8, "dragon", "", "", ""},
	},
	"scarlet-violet": {
		{"katy", "cortondo-gym", 1, "bug", "", "", noteFreeOrder},
		{"brassius", "artazon-gym", 2, "grass", "", "", noteFreeOrder},
		{"iono", "levincia-gym", 3, "electric", "", "", noteFreeOrder},
		{"kofu", "cascarrafa-gym", 4, "water", "", "", noteFreeOrder},
		{"larry", "medali-gym", 5, "normal", "", "", noteFreeOrder},
		{"ryme", "montenevera-gym", 6, "ghost", "", "", noteFreeOrder},
		{"tulip", "alfornada-gym", 7, "psychic", "", "", noteFreeOrder},
		{"grusha", "glaseado-gym", 8, "ice", "", "", noteFreeOrder},
	},
}

// Orden cronológico de procesamiento.
var gameOrder = []string{
	"red-blue", "yellow", "gold-silver", "crystal", "ruby-sapphire", "emerald", "firered-leafgreen",
	"diamond-pearl", "platinum", "heartgold-soulsilver", "black-white", "black-2-white-2", "x-y",
	"omega-ruby-alpha-sapphire", "lets-go-pikachu-lets-go-eevee", "sword-shield",
	"brilliant-diamond-shining-pearl", "scarlet-violet",
}
