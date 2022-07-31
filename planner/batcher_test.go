package planner

import (
	"strings"
	"testing"
)

func TestStreamBatcherChatterSizeLowerThanMax(t *testing.T) {
	t.Parallel()

	const chatterCount = 3
	b := &StreamBatcher{
		MaxQueueSize: 10,
		FlushFunc:    func(queue []string) {},
	}
	b.ChatterSize = chatterCount

	if b.queue != nil {
		t.Fatal("expected queue to be empty")
	}

	b.Enqueue("user1")
	gotl, gotc := len(b.queue), cap(b.queue)
	wantl, wantc := 1, chatterCount
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 2, chatterCount
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// should trigger flush
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}

	if b.flushCount != 1 {
		t.Fatalf("expected 1 flush, got %d", b.flushCount)
	}
	if b.queueCount != 0 {
		t.Fatalf("expected queueCount to be 0. No more flushes needed")
	}
}

func TestStreamBatcherChatterSizeZero(t *testing.T) {
	t.Parallel()

	const chatterCount = 0
	const max = 5
	b := &StreamBatcher{
		MaxQueueSize: max,
		FlushFunc:    func(queue []string) {},
	}

	if b.queue != nil {
		t.Fatal("expected queue to be empty")
	}

	b.Enqueue("user1")
	gotl, gotc := len(b.queue), cap(b.queue)
	wantl, wantc := 1, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 2, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 3, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}

	if b.flushCount != 0 {
		t.Fatalf("expected 0 flush, got %d", b.flushCount)
	}
	if b.queueCount != 3 {
		t.Fatal("expected queueCount to be 3, we should have 3 items to be flushed")
	}
}

func TestStreamBatcherChatterSizeZeroMoreValuesThanMax(t *testing.T) {
	t.Parallel()

	const chatterCount = 0
	const max = 3
	b := &StreamBatcher{
		MaxQueueSize: max,
		FlushFunc:    func(queue []string) {},
	}

	if b.queue != nil {
		t.Fatal("expected queue to be empty")
	}

	b.Enqueue("user1")
	gotl, gotc := len(b.queue), cap(b.queue)
	wantl, wantc := 1, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 2, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// Should trigger flush
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 1, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 2, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}

	if b.flushCount != 1 {
		t.Fatalf("expected 1 flush, got %d", b.flushCount)
	}
	if b.queueCount != 2 {
		t.Fatal("expected queueCount to be 2, we should have 2 items to be flushed")
	}
}

func TestStreamBatcherChatterSizeSameAsMax(t *testing.T) {
	t.Parallel()

	const chatterCount = 3
	const max = 3
	b := &StreamBatcher{
		MaxQueueSize: max,
		FlushFunc:    func(queue []string) {},
	}
	b.ChatterSize = chatterCount

	if b.queue != nil {
		t.Fatal("expected queue to be empty")
	}

	b.Enqueue("user1")
	gotl, gotc := len(b.queue), cap(b.queue)
	wantl, wantc := 1, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 2, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// Should never trigger flush. If ChatterSize = 0 we can't predict size nor
	// flushes
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}

	if b.flushCount != 1 {
		t.Fatalf("expected 1 flush, got %d", b.flushCount)
	}
	if b.queueCount != 0 {
		t.Fatalf("expected queueCount to be 0. No more flushes needed")
	}
}

func TestStreamBatcherChatterSizeGreaterThanMax(t *testing.T) {
	t.Parallel()

	const chatterCount = 8
	const max = 3
	b := &StreamBatcher{
		MaxQueueSize: max,
		FlushFunc:    func(queue []string) {},
	}
	b.ChatterSize = chatterCount

	if b.queue != nil {
		t.Fatal("expected queue to be empty")
	}

	b.Enqueue("user1")
	gotl, gotc := len(b.queue), cap(b.queue)
	wantl, wantc := 1, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 2, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// Should trigger flush
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 1, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 2, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// Should trigger flush
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	// Should know that there is only 1 element left and allocate with that
	// element size instead of max
	wantl, wantc = 1, 2
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// Should trigger flush
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}

	if b.flushCount != 3 {
		t.Fatalf("expected 3 flush, got %d", b.flushCount)
	}
	if b.queueCount != 0 {
		t.Fatalf("expected queueCount to be 0. No more flushes needed")
	}
}

func TestStreamBatcherChatterSizeWrong(t *testing.T) {
	t.Parallel()

	const chatterCount = 3
	const max = 2
	b := &StreamBatcher{
		MaxQueueSize: max,
		FlushFunc:    func(queue []string) {},
	}
	b.ChatterSize = chatterCount

	if b.queue != nil {
		t.Fatal("expected queue to be empty")
	}

	b.Enqueue("user1")
	gotl, gotc := len(b.queue), cap(b.queue)
	wantl, wantc := 1, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// Should trigger flush
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// Should trigger flush
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}

	// We passed a ChatterSize = 3 but we will enqueue 5 elements
	// This should trigger a new cycle with MaxQueueSize=2
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 1, max
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}
	// Should trigger flush
	b.Enqueue("user1")
	gotl, gotc = len(b.queue), cap(b.queue)
	wantl, wantc = 0, 0
	if gotl != wantl || gotc != wantc {
		t.Fatalf("got len:%d cap:%d\nwant len:%d cap:%d", gotl, gotc, wantl, wantc)
	}

	if b.flushCount != 3 {
		t.Fatalf("expected 3 flush, got %d", b.flushCount)
	}
	if b.queueCount != 0 {
		t.Fatalf("expected queueCount to be 0. No more flushes needed")
	}
}

func TestStreamBatcher(t *testing.T) {
	obj := strings.NewReader(`"{"_links":{},"chatter_count":1204,"chatters":{"broadcaster":["polispol1"],"vips":["ariian_amy","noquemecansus"],"moderators":["agustin838","cabruu","d0oppa","monguerra","oscarval16","srmauri_","streamelements"],"staff":[],"admins":[],"global_mods":[],"viewers":["0_0zer0_","0_ayanami_0","0_nekopara_0","1norngs","1ruvik3","1ryan_378","21skyblue","4chappy","4lejo_j","504_brayan","64nanete","7th_dummy","8daftpunk8","934casper","95rayomcqueen","9dil3","a13_r","a_carlos_a","a_x_e_l_c","aantonic77","aarantxa7","abdalone10","abriperez","acubach","adamary3","adeunx","adgamboa24","adri2301","adrian211977","adrielda","adrigt98","adriiann113","adriln99","adrisagar","aekyl","againtothefuture","agelandrosp","aguadecoco2020","aguskyl","ajm1781992","akasalem","alaimary","alangrizz","albaaasf","albacat","albertmartorell7","aldebaran019","aldrack1990","ale_yupa","alebenaduss","alecoceres","alecpfx","alegarcia67","alejandro10196","alesa75","alessandra_6","alexduone","alexisonthesauce","alexisorrego1","alexlebrn","alexs_jazz","alexxx_2205","alficonfanta","alfredovicens","alguiendelm","ali_relok4","aliceydra","aliciagcz","alier3d","alixhuerta","alizeeku","almaria89","almudena_ig","alonsby456","aluech","alvarito1494","alvaro9000","alvaroglez19942","alwickk","amebasriper","amegun","amn_5","an49135","ana505beka","anahio_o","anapple4","anarolve","anarupu","ancorvantian1","andonix29__","andrea_br_30","andreaa__17","andreatorresv","andreav1a","andres2394","andres46564","andresvamp","andresxks","andreusdark","andyyy_pz","angelgzzs","angelito47","angelkaido13","angellojs","angelpintxos","angi373","angienicoled","annieleath","annita675","annorlondo","anonimo___","anotherttvviewer","antimadrugista","antomms","antuiogames","anyelayisse","apcaparros","aragon_84","aranxiitaa","arcimboldo_boldo","ari_stella","ariel5794","arimon_03","arimonsmtz","aritzua","arkinod_","arleysandoval07","armentta","arnausanche","artaixde","arteida","artur_gomariz","asdrubaljrrr","aslozano","assengard","astur004","atek46019","aten","atomdeid","autumnbag153338","avyanna90","axelperdomoxd","axrodri1","ayla_74","ayllenpl","aymi_22","ayusooooo","azucarados_chocolatosos","b_matus","barbirocio","barracus098","barulleiro","belen_ichigo","beleniuss","benisor","benlli_09","bestbusybee","biagioguinea79","billcapi13","biomess_","blackansito","bladecore","blancabcn","blaskarenrique","bluearrow_208","boern76","boiser","bombita2","boqueroncetepc03","borjafender","boty28","brayan99_hm","brealing10","brocoly44","brujomatias","bryan_ro22","bubu_dpa","bullriver13","buore","bylilu","bysierra25","c_pastor","caalvario","caandelaa_grm","caatalinapaz","calabera112","calaviadj","caleb_214","cam_kaos23","camimm23","caotictaia","carlaakar","carlagonzzza","carlaguiilarv","carlbra17","carlesfu","carlespo","carlos_g2703","carlos_ortiz7","carlosperezb","carlosruizzaragoza","carly_14","carmelitamor","carmen_boanerges","carobregli","carolinasusana","cassiusklein","cbella31","cecilio_gf","celes_y05","cesc11","cfclub17","cgzmn11","chakyown","chebbito12","cheldon13","chesardo","chicochilll","chicua111","chiin222","chitixcx","choklo79","chompoff94","chrisvelasquez_","churre18","ciscanito1973","cisf15","claris_monis","clau0806","claudia190306","claudiarivero","claudiomoraga","cleisonca","clipandcloper","clopez_99","colo_aria30","commanderroot","conjdejaime","copycaaaat","cosas_chingonas","crazelty","criisgrande","crisescu","crisquetglas97","cristianfcm","cristo629","cristty991","crokorox","croquialba","cxrlitos99","daaviidgzz","dabeatdj","dabstrax","daher_g","daigoplays","daisylau14","dalayus","dana_darko","danichimp_","daniee11","danieeeeeel13","danieiita21","daniel020202020202","daniel7dh","danieldduran","danielfe_95","danigagui","daniploplo","danirama_","danito8_","dannylc1","dannynn","dannypirate_78","dannytovar1221","danted_20","dardignac","dari_flores","dario_23s","darktrivi","davidcaraballobulnes","davidcf2","davidcollado18","davidexter","davidrec_","davidsm83","davidvlnc","davizhoo","daxteryo","deavitt_","deboritiswoman","decalcom_anya","denisebarto","denisgod555","descriteriadx","destcraun","dianabrangel","diegoalvarez95","diegoflorez18","diegogb93","diegolc78","diegoo_uwu","diiianagomez","dilemma928","dinosairuz","dioni_14","dix_ragnvindr","djang0v","djmurci1","djxavi95","docjj76","docks23","dolfann1996","donbugui","donchiti_13","dondropex","donius_","donjofe","dragonplay022","dragonroj13","drakupsn","dsith","dst_dop3","duffyporta","dunki3","dvlpr1","dylanrocherr","eabastida21","ebo2933","eddysjo","eduardommt98","edukrampus","edydelcid110","egarrik","egrl196","egv96","elazotedi0s","elbruna37","elcalvoivan","elchachotwain","elchuma0","elcitim","electriccrazypr","elenaac_19","elenamont_","elflecha1","elgeneral0401","elidrasil","elii_aravena","elisma69","eljiren_","eljoze2009","elkaod","elkirby115","ellieemorales","elpersienaspadillo","elpjota","elrincondelarelop","elschutze","eltareas","elturu9","emi32river","emiliolastra","emilse_lera","enderchast20_03_05","enmasaga","eosszett","eric_v95","eriklajas","erre__","ertit410","esconejo","eslilian_rp","esqueletoo_o","esterler2","eugipeart","eziomarley","fab14n_miranda","fabiogm8","fabsgm","fallguysak","fantasmiito","farbautis","fatifg","febritt","fedelandriel88","feeernino","felipe_1605","ferb2902","ferbolagnos","fermassa1981","fernandortiz02","fisik0ne7","fl0orencia","flem1t444","florezz_12","florido","fontyed","francosm20","frank_42_","franwillo","freaktoad","fredysantos03","frikiluser","fsguevara","fuguixx","fulcrum013","g2chris","gabbcanm","gabo_rv21","gabytere","galicia2307","gandreel","gapavela","garinga1984","gary_efron","gemabus","geodonz","gexerx","ghost_fsc","gilit30","gisuwu23","gliqui","glorsssss","gluglujaja","gonmon25","gonza2525","gonzalofn2008","griraspu","guaje0707","guillermofs33","gutierrxz_marina","gutijudo","h6alth","h_d_hades","hackros","hadtor","halberto24","hanapuma999","hanniwistriquis","harryscamander881","haxox31","hazeljameson","hect0or_","helensm2009","henz_17","herefrom1979","hergregory","hermes_prime","hiruma_hernandez","honoka_312","huggy_playtime_","hugo_sc03","hugoman2401","hulugra","huntermen99","iamjeanjy","iamricardoramz","iamrocky25","ian_hernandez27","igoor_valdes","iharry31","iillzerollii","ikercp","ilovekoiuwu","im_adler","im_l3mon","imariajosei","inmadvegman","intrhax_angello","ioannesmagnus","ioazuo","iree_sirgo","irisbolmaro","iro_n_algas","iron666cat","ironfigther","ironwoman81","is1ote","isa2345667","isaac_espi","isaizquierdo1298","isalmela","ishirocks","iskra098","itachii2022","itjhyjuo","itsabraham1706","itsruih_8","ivan198810","ivan_kvn","ivan_uru9","ivanapumcake","ivi76racing","ivourinho","ivy03p","izydron","jackberman04","jackmarcelo666","jacob_scz","jaider4678","jairsito222_","jakhed","jalferrari","jani_sepu","jare_murga","jariel1599","jary_1978","jasperit0","javi_moreno","javicollarte","javier_calvo97","javier_marcos_","javierhappy01","javigarcia_07","javilarod","javimarmota","javin3211","javitowp","javivm92","javo1928","javu70","jazmin_gq","jcormas80","jelensita","jenisanz11","jenni920_pr","jennloli","jennsz","jeremiasbc","jeremy_1098","jericknm","jeslynsamar","jesus_sanchez_11","jesuspar90","jgagahpa","jheremy_25","jijicat777","jivannunez_1017","jjohan_r_","jjsalass","jldomin","jmboan","jnico1096","joaco_i","joakiferlindo","joan_wp","joanbe2","jobaga_","jocmyk","joelgarciia_","joelius28","johntonic1","joinus710","jome_710","jonasalmon","jonelo34","jonir1903","jonsu1313","jooselinne","jor9k_","jorgebest75","jorgecrc","jorgegarciarosales","jorgex_55","josa011","jose_0132","jose_carlis_84","joseafk10","josealdo14","josegogo13","joselito1987","joseps_01","josuegr31","josuera26","jsantos996","juaanxe","juampya","juancarlitos46","juancavaglia","judithnnavarro","jug07","juggertbolt","juliaaaa2004","juliacr99","juliovegaconh","justbeong","jy3sus","jyaiy09","kairobe98","kaiservidal","kakadasilv","kalimist72","kamus83","kancersito","kapahala26","kassadd","katkit","katrik_0","kats_154","katya_psic_forense","katzeniv","kelvedar","kepwoco","kerbe5","keviin93","kevinborniard2","kevinlire","khanmaster239","khristo0987","kier0tu","kikeewash","killerqueen313","kimi3headsmonkey","kingstar_91","kitkat0000","koagameesp","koalitac","koke_explosive","kolokon96","kozyeleven","krisis_jl","kufat1","kurisutkr","labosirk","ladiegaaaaa","lafalu08","lagorradeputu","lajaibamordelona","laloupez","lamamarexulona","lamoishe","lamstherd","lancary","lancito_5","lanurinuri23","laprimadeari","lau1198","lauloves27","laura_abeto","laura_coune","laura_sv_","laurabloodflowerssmith","lauryemyers","laznaitt5","lebusin","leireesi","leli_h","leonare2","lestico42","leto89atreides","leumaseb","levysan_3","lex38b","leyenblack","liamc2000_","lichl_","lidia_gualda","lilagus27","lilymap","lina2845","lira196","lisaliss16","little_revi","littlejulia98","lizeeth","lizlobita","llanerito","lmpep","logicalhardware","lokis80","lolailaura","lolop88y_","lomitooooooooo","lord_oscuro","lorenarodriguezval","lorowi","losmasdeportivos","losprofesdeciencias","loubegp98","lucasrhymer","luccylynn","luciaapsancheez","luh_xnsth","luis_valencia30","luisfuentes1510","luishellsing2","luisjrivas","luislobo_22","luisma_carbayon","luismexxi","luissofipi","maaaaaatttttttt","maascotas","macam_m","mafe_368","magjsr","magnetron223","maku188","mandarinafucsia","maneiroxd","manoelhernandez","manuelbomu","manug714","mar5cos5","marajf","marcecee","marcl5","marcomtrm","marcrovirapuchol","margaq","mari0_31","mari_loli4","maria070309","maria_mh00","mariale_calu","marialerrodriguez","marialopezr","mariana1255","marianolrc","mariasanz621","marietaps","marika_confi","marinita_08","mario_18q","mariofruc","marioo__19","markitus97","marliiinee","marta0735","marta_fv","marta_vf99","martaanillo","martarive17","martinbuitra05","martinrp429","martitathera","marvin_2121","mary_m03","marymq","mati_rulin","matimugica","matiparolari","matteo3312","matthew_13087","mauconzeta","mauxd24","maylareina","mcmike013","meikusmc","mekusarg","melanicn","melideni25","meryyomisma","mhispano","miauh24","michiangels","miguel_minipunk","miguel_mtzz","mikemagician25","miriam__san","misarpe","missotoe","misteravi_","misterberete","mistz","mixfrutal","miyatemiyamoto","mmoran5580","mndz044","mohicanokush","moira_37","molsito_","mondpulver","monicafp","monicaruizz","montsearcoss","monttblack094","moobot","mordeadiestramientocanino","motenai77","moxkor2020","moya242","mr_cva","mrtable07","mrtripleeeee","msmarvel87","muerto_d_hambre","murillo_099","mutekkk","mymmoo","mynor_osorio","mysstogan","n0e97","nachig20","nachitomartinez0712","nachobals","nahiatxu_u","nallelyluna06","namelesslor","nanakirydia","nanasnake","naniibp","nanuk93__","nany28","napraxd","nataliaasnchez","natysapoconcho","nay_ie","nayelycolz1002","nayrb_alejandro","nazarethortega","neilith8","nekosnow02","nelsonglife","nelsymejiia","nereeaa29","nesukochaaan","netho_11","newyamakasy","nhoyiro","nhyonz","niam009","nic0j","nicolasgm_","nicolaslizardo44","nixcitoo","noa_fufi","noa_tr","noe_787","noelia2199","noesantalla","noligonunca","non4me221","none_elsilla","nrggg","nulla_033","nuria_twd","nuriaibma92","o_black7","ohveyes","oijkio_420","olga1020","olisoyfran","olivermelgarr","omi744","omongus1999","oskkun","osquitar1985","oswaldodvl12","ouchurusconcara","p0yolion","paaatriii_","paae11","paatroo22","pablisimos","pablobemar5","pachu_37","palmeta_74","pamisaurr","panzonk","papijvancho","pascuu9","patitomcpato","patricku","pau9daniel","paukixd","paulaaleeon_","pauly_24","pax_451","pedrodc_26","peremat","perilipo","perrrosarnossso","perry0609","pertivana","pete_el_mediano","petoforce37","phantom__44","piccolojr_","pickolin","piero120996","pieropino10","pikachu_up","pikukoide","pintofrey","piratadejuguete","pl0m1t0","playerone_glhf","policemau","polleeeeeeeete","pollo31","polmarin06","ponchozim","porteraz0200","pouyiojs","pqtrng_27","proto_c","psc008","psichicc","punketus","puurpa","qaaqster","qfg_yaranaika","qsy12344","quazarts","queryselectorall","quimy_fer","quintaniya","quiqueortez","quiqueperez14","quoziur","raachuu86","raaqueee","racsody","radioauchi","rafa_psique_","rafosl0194","raizerotnigzz","rakipy","ramsalija","raquelevr","raul_bx","raulzh78","raydezel","reaper_soul20","regulus_99","renskar","renzo365","resilencej","respeto_72","reyenriquee","rherreroe","riawe_","rich_ltc","richinoguera","rj19982","ro_rivallie","robinsso","rocio_136","rocio_lozanoo","rodri15_","rodri2324","rodrigosanchez77","rodriguez_1312","rojillo33vk","romansoutoo","romanvm1","romiswiftie","ron2k2","ronrolove","roxvan","rua936","rubenpg_20","rubioikling","rucius2","rulisraulmartinez","rusty_890_","ruthlpzt","ryakk_","rysenn_","s0opa4a","sabru__","sachemosz","saez88","sam_fxc","samuelaguilar22","samuuuu178","sandy952","santew1","santi2047","santiagogn","santicrqkcxs","sara23m","saramvr11","sayuri_olga","scion_31","sdel95","seb0019","sebacar75","sebarl_8","sebas_dj_","sebastian213114","sebastiangc97","sebastianstaion4","secalca","segadream","sejulian","sembrando_la_yuca_","senda_17","sendeshi180","seneka_","seraagal","serchiomg","sergi_fr","sergicop","sergigf05","sergio9hdez","sergioyasiis85","sernorui","sh1ranui_","shadowfell969","shadowsyzer","shaqtar78","sharksbd69","shinda_13","shiro06e","shomi08","shyryuddaeng","sid_kai","sidrmzam","silentlurker998","silvina_laura","sintiacryo","skillerpv_yt","skruul_75","soalxz_","sofiagarz0n","sofiamtnez08","sofiatorresmorales","sofibailetti","solracxxd","sombraways","sonally_26","sonne_314","sophikal","soydemando_644","soyfresita_","soygermain","soygo_loso","specialone1984","spiketrapclair","spreenistass","spriinxd","spritecarr","sra_wikita","srdteach","srjupiter9","srkaneki55","srodriguez2043","srpate__","ssojod14","starkkk___","stebant","stephcz95","strikerpow","su_sana_horia_play","suguru78","suiza31","suprador_","supremogana1508","surimara","suso1988","sybilla_bathory","t0mako","tahielm","tailerdurden79","takamas1902","tam_baleo","taniasb","tardar_sauce","taroku15","tarokun18","tatsulkuirel","tbvtori","tekitolalife","teofania69","tequila10k","teranne","terybully","thaneissa","thebordotv","thejxlioo","thekreatorred","theotakugrl","thepaulpetershow","thesergi_10","theshadowxtv","thiagosc3","thiax11","thoronres","titina2605","tiwar","tizcloud","tkdndifnd","toni1024","tony_lfx","tootti_1402","toxic_ramii","tp_cattiol","trizrianoxx","trooiitii","truenostradamus","trujy87","ttysofia","tuca74","turambare","turip_125ipip","twls0","txopo2317","ubemd","unafriki","unaiar06","uncafeconletras","unlimit3d_py","unpinwino","uxiavila","v12345ca","vainestbump15","vainilla_vainilla","valentin_vy","valeria_beccerril_diaz","valldeim_56","vanahimm","varoc12__","vathonj","venbeng","veronii_ca","veroverito03","vicbaza","vicentecs_","vicfalk","victoriaaledo","viernes_trece","vikerful","vikingmonkeygamer","viktorcollantes","viky_146","vilieslock","villagebread","virginia_leo","virgispm","visabis","visvii6_","vita_chamber","wackycheese22","waldecktbd","wanacoar","wardtrigger","weedom","wesftw","willitro","wilman32645","wilmarsilva_","wolfhellsing2","worfworfspspsp","x_tania","xaleaguilar","xavierdemons","xavigar10","xenouii","xgaby_05x","xharlymp","xiajcm","xiquilla","xjeok","xkeshaxx","xkunma","xoxxkanekixxox","xtinchisx","xx_aragon_xx","xxalphalyonxx","xxginx2041xx","xxmadaraxxcc","xxnicolas111xx","xxred720xx","xxshadowforcexx0","yaaiizeen","yahiara159","yamir_unu","yaneira_07","yassiinz","yekgian","yeray170","ygbasense","yoen_long","yohananfdzz","yonathancanes1996","yopegfx","yosep1980","yosoybobi","yossj27","young3lood","yuriii_saan","yuyu__20_10_04","zanixuss","zekyov","zeusipo","zobala66","zoebelle12","zoomlaserpewpew","zosser101","zst1santi","zumbalemambo","zvc_fl"]}}"`)

	b := NewStreamBatcher()
	b.Batch(obj)

}
