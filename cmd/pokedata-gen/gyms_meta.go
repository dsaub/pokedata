package main

// Metadatos de líderes y gimnasios que no vienen en las fuentes de equipos.

type leaderDef struct{ id, en, es, gender string }

var leaderDefs = []leaderDef{
	// Kanto
	{"brock", "Brock", "Brock", "M"},
	{"misty", "Misty", "Misty", "F"},
	{"lt-surge", "Lt. Surge", "Lt. Surge", "M"},
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
	{"crasher-wake", "Crasher Wake", "Mananti", "M"},
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
	{"pewter-gym", "Pewter City", "pewter-city", "kanto", "Boulder Badge"},
	{"cerulean-gym", "Cerulean City", "cerulean-city", "kanto", "Cascade Badge"},
	{"vermilion-gym", "Vermilion City", "vermilion-city", "kanto", "Thunder Badge"},
	{"celadon-gym", "Celadon City", "celadon-city", "kanto", "Rainbow Badge"},
	{"fuchsia-gym", "Fuchsia City", "fuchsia-city", "kanto", "Soul Badge"},
	{"saffron-gym", "Saffron City", "saffron-city", "kanto", "Marsh Badge"},
	{"cinnabar-gym", "Cinnabar Island", "cinnabar-island", "kanto", "Volcano Badge"},
	{"viridian-gym", "Viridian City", "viridian-city", "kanto", "Earth Badge"},

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

	{"oreburgh-gym", "Oreburgh City", "oreburgh-city", "sinnoh", "Coal Badge"},
	{"eterna-gym", "Eterna City", "eterna-city", "sinnoh", "Forest Badge"},
	{"veilstone-gym", "Veilstone City", "veilstone-city", "sinnoh", "Cobble Badge"},
	{"pastoria-gym", "Pastoria City", "pastoria-city", "sinnoh", "Fen Badge"},
	{"hearthome-gym", "Hearthome City", "hearthome-city", "sinnoh", "Relic Badge"},
	{"canalave-gym", "Canalave City", "canalave-city", "sinnoh", "Mine Badge"},
	{"snowpoint-gym", "Snowpoint City", "snowpoint-city", "sinnoh", "Icicle Badge"},
	{"sunyshore-gym", "Sunyshore City", "sunyshore-city", "sinnoh", "Beacon Badge"},

	{"striaton-gym", "Striaton City", "striaton-city", "unova", "Trio Badge"},
	{"nacrene-gym", "Nacrene City", "nacrene-city", "unova", "Basic Badge"},
	{"castelia-gym", "Castelia City", "castelia-city", "unova", "Insect Badge"},
	{"nimbasa-gym", "Nimbasa City", "nimbasa-city", "unova", "Bolt Badge"},
	{"driftveil-gym", "Driftveil City", "driftveil-city", "unova", "Quake Badge"},
	{"mistralton-gym", "Mistralton City", "mistralton-city", "unova", "Jet Badge"},
	{"icirrus-gym", "Icirrus City", "icirrus-city", "unova", "Freeze Badge"},
	{"opelucid-gym", "Opelucid City", "opelucid-city", "unova", "Legend Badge"},
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
	{"stow-on-side-stadium", "Stow-on-Side", "stow-on-side", "galar", "Fighting Badge"},
	{"ballonlea-stadium", "Ballonlea", "ballonlea", "galar", "Fairy Badge"},
	{"circhester-stadium", "Circhester", "circhester", "galar", "Rock Badge"},
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
	{"brock", "pewter-gym", 1, "rock", "", "", ""},
	{"misty", "cerulean-gym", 2, "water", "", "", ""},
	{"lt-surge", "vermilion-gym", 3, "electric", "", "", ""},
	{"erika", "celadon-gym", 4, "grass", "", "", ""},
	{"koga", "fuchsia-gym", 5, "poison", "", "", ""},
	{"sabrina", "saffron-gym", 6, "psychic", "", "", ""},
	{"blaine", "cinnabar-gym", 7, "fire", "", "", ""},
	{"giovanni", "viridian-gym", 8, "ground", "", "", ""},
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
	{"brock", "pewter-gym", 9, "rock", "", "", "Kanto (post-game)"},
	{"misty", "cerulean-gym", 10, "water", "", "", "Kanto (post-game)"},
	{"lt-surge", "vermilion-gym", 11, "electric", "", "", "Kanto (post-game)"},
	{"erika", "celadon-gym", 12, "grass", "", "", "Kanto (post-game)"},
	{"janine", "fuchsia-gym", 13, "poison", "", "", "Kanto (post-game)"},
	{"sabrina", "saffron-gym", 14, "psychic", "", "", "Kanto (post-game)"},
	{"blaine", "cinnabar-gym", 15, "fire", "", "", "Kanto (post-game); el gimnasio está en las Islas Espuma"},
	{"blue", "viridian-gym", 16, "", "", "", "Kanto (post-game)"},
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
	{"roark", "oreburgh-gym", 1, "rock", "", "", ""},
	{"gardenia", "eterna-gym", 2, "grass", "", "", ""},
	{"maylene", "veilstone-gym", 3, "fighting", "", "", ""},
	{"crasher-wake", "pastoria-gym", 4, "water", "", "", ""},
	{"fantina", "hearthome-gym", 5, "ghost", "", "", ""},
	{"byron", "canalave-gym", 6, "steel", "", "", ""},
	{"candice", "snowpoint-gym", 7, "ice", "", "", ""},
	{"volkner", "sunyshore-gym", 8, "electric", "", "", ""},
}

var ptSlots = []slotDef{
	{"roark", "oreburgh-gym", 1, "rock", "", "", ""},
	{"gardenia", "eterna-gym", 2, "grass", "", "", ""},
	{"fantina", "hearthome-gym", 3, "ghost", "", "", ""},
	{"maylene", "veilstone-gym", 4, "fighting", "", "", ""},
	{"crasher-wake", "pastoria-gym", 5, "water", "", "", ""},
	{"byron", "canalave-gym", 6, "steel", "", "", ""},
	{"candice", "snowpoint-gym", 7, "ice", "", "", ""},
	{"volkner", "sunyshore-gym", 8, "electric", "", "", ""},
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
		{"chili", "striaton-gym", 1, "fire", "", "", "Si el jugador eligió a Snivy"},
		{"cress", "striaton-gym", 1, "water", "", "", "Si el jugador eligió a Tepig"},
		{"cilan", "striaton-gym", 1, "grass", "", "", "Si el jugador eligió a Oshawott"},
		{"lenora", "nacrene-gym", 2, "normal", "", "", ""},
		{"burgh", "castelia-gym", 3, "bug", "", "", ""},
		{"elesa", "nimbasa-gym", 4, "electric", "", "", ""},
		{"clay", "driftveil-gym", 5, "ground", "", "", ""},
		{"skyla", "mistralton-gym", 6, "flying", "", "", ""},
		{"brycen", "icirrus-gym", 7, "ice", "", "", ""},
		{"drayden", "opelucid-gym", 8, "dragon", "black", "", ""},
		{"iris", "opelucid-gym", 8, "dragon", "white", "", ""},
	},
	"black-2-white-2": {
		{"cheren", "aspertia-gym", 1, "normal", "", "", ""},
		{"roxie", "virbank-gym", 2, "poison", "", "", ""},
		{"burgh", "castelia-gym", 3, "bug", "", "", ""},
		{"elesa", "nimbasa-gym", 4, "electric", "", "", ""},
		{"clay", "driftveil-gym", 5, "ground", "", "", ""},
		{"skyla", "mistralton-gym", 6, "flying", "", "", ""},
		{"drayden", "opelucid-gym", 7, "dragon", "", "", ""},
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
		{"bea", "stow-on-side-stadium", 4, "fighting", "sword", "Fighting Badge", ""},
		{"allister", "stow-on-side-stadium", 4, "ghost", "shield", "Ghost Badge", ""},
		{"opal", "ballonlea-stadium", 5, "fairy", "", "", ""},
		{"gordie", "circhester-stadium", 6, "rock", "sword", "Rock Badge", ""},
		{"melony", "circhester-stadium", 6, "ice", "shield", "Ice Badge", ""},
		{"piers", "spikemuth-gym", 7, "dark", "", "", ""},
		{"raihan", "hammerlocke-stadium", 8, "dragon", "", "", ""},
	},
	"scarlet-violet": {
		{"katy", "cortondo-gym", 1, "bug", "", "", "Orden libre; orden por nivel"},
		{"brassius", "artazon-gym", 2, "grass", "", "", "Orden libre; orden por nivel"},
		{"iono", "levincia-gym", 3, "electric", "", "", "Orden libre; orden por nivel"},
		{"kofu", "cascarrafa-gym", 4, "water", "", "", "Orden libre; orden por nivel"},
		{"larry", "medali-gym", 5, "normal", "", "", "Orden libre; orden por nivel"},
		{"ryme", "montenevera-gym", 6, "ghost", "", "", "Orden libre; orden por nivel"},
		{"tulip", "alfornada-gym", 7, "psychic", "", "", "Orden libre; orden por nivel"},
		{"grusha", "glaseado-gym", 8, "ice", "", "", "Orden libre; orden por nivel"},
	},
}

// Orden cronológico de procesamiento.
var gameOrder = []string{
	"red-blue", "yellow", "gold-silver", "crystal", "ruby-sapphire", "emerald", "firered-leafgreen",
	"diamond-pearl", "platinum", "heartgold-soulsilver", "black-white", "black-2-white-2", "x-y",
	"omega-ruby-alpha-sapphire", "lets-go-pikachu-lets-go-eevee", "sword-shield",
	"brilliant-diamond-shining-pearl", "scarlet-violet",
}
