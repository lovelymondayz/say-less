package matcher

import (
	"strings"
)

// Indonesian to English translation dictionary for common words
var indonesianDict = map[string]string{
	"AKU": "I", "SAYA": "I", "KAMU": "YOU", "DIA": "HE", "MEREKA": "THEY",
	"KITA": "WE", "KAMI": "WE", "INI": "THIS", "ITU": "THAT",
	"SUKA": "LIKE", "CINTA": "LOVE", "BENCI": "HATE", "MAU": "WANT",
	"BISA": "CAN", "HARUS": "MUST", "ADA": "THERE", "PERGI": "GO",
	"PULANG": "HOME", "DATANG": "COME", "LIHAT": "SEE", "DENGAR": "HEAR",
	"MAKAN": "EAT", "MINUM": "DRINK", "TIDUR": "SLEEP", "BANGUN": "WAKE",
	"KERJA": "WORK", "BACA": "READ", "TULIS": "WRITE", "NAIK": "UP",
	"TURUN": "DOWN", "MASUK": "IN", "KELUAR": "OUT", "BUKA": "OPEN",
	"TUTUP": "CLOSE", "BERI": "GIVE", "AMBIL": "TAKE", "BELI": "BUY",
	"JUAL": "SELL", "BAYAR": "PAY", "KIRIM": "SEND", "TERIMA": "RECEIVE",
	"TANYA": "ASK", "JAWAB": "ANSWER", "PIKIR": "THINK", "TAHU": "KNOW",
	"INGAT": "REMEMBER", "LUPA": "FORGET", "COBA": "TRY", "MOHON": "PLEASE",
	"KASIH": "GIVE", "BANTU": "HELP", "TOLONG": "HELP",
	"TUNGGU": "WAIT", "CEPAT": "FAST", "LAMBAT": "SLOW",
	"BESAR": "BIG", "KECIL": "SMALL", "PANJANG": "LONG",
	"PENDEK": "SHORT", "TINGGI": "TALL", "BERAT": "HEAVY", "RINGAN": "LIGHT",
	"BAGUS": "GOOD", "JELEK": "BAD", "INDAH": "BEAUTIFUL", "CANTIK": "PRETTY",
	"GANTENG": "HANDSOME", "PINTAR": "SMART", "BODOH": "STUPID", "KUAT": "STRONG",
	"LEMAH": "WEAK", "PANAS": "HOT", "DINGIN": "COLD", "BASAH": "WET",
	"KERING": "DRY", "PENUH": "FULL", "KOSONG": "EMPTY", "MUDAH": "EASY",
	"SUSAH": "HARD", "MURAH": "CHEAP", "MAHAL": "EXPENSIVE", "GRATIS": "FREE",
	"BARU": "NEW", "LAMA": "OLD", "PERTAMA": "FIRST", "TERAKHIR": "LAST",
	"SEKARANG": "NOW", "KEMUDIAN": "THEN", "NANTI": "LATER", "TADI": "EARLY",
	"TERLAMBAT": "LATE", "SELALU": "ALWAYS", "TIDAK": "NOT", "TDK": "NOT",
	"YA": "YES", "MUNGKIN": "MAYBE", "BANYAK": "MUCH", "SEDIKIT": "LITTLE",
	"LEBIH": "MORE", "KURANG": "LESS", "PALING": "MOST",
	"SEBELUM": "BEFORE", "SESUDAH": "AFTER", "SELAMA": "DURING", "TANPA": "WITHOUT",
	"DENGAN": "WITH", "UNTUK": "FOR", "DARI": "FROM", "KE": "TO", "DI": "AT",
	"PADA": "ON", "DALAM": "IN", "LUAR": "OUT", "ATAS": "ABOVE", "BAWAH": "BELOW",
	"DEPAN": "FRONT", "BELAKANG": "BACK", "SAMPAI": "UNTIL", "ANTARA": "BETWEEN",
	"ADALAH": "IS", "MENJADI": "BECOME", "MEMPUNYAI": "HAVE", "MENGGUNAKAN": "USE",
	"SENANG": "HAPPY", "SEDIN": "SAD", "MARAH": "ANGRY",
	"TAKUT": "SCARED", "CEMBURU": "JEALOUS", "MALU": "SHY", "BANGGA": "PROUD",
	"KECEWA": "DISAPPOINTED", "BINGUNG": "CONFUSED", "LELAH": "TIRED", "SEHAT": "HEALTHY",
	"SAKIT": "SICK", "LAPAR": "HUNGRY", "HAUS": "THIRSTY", "PUSING": "DIZZY",
	"SAHABAT": "FRIEND", "KELUARGA": "FAMILY", "AYAH": "FATHER", "IBU": "MOTHER",
	"ANAK": "CHILD", "ADIK": "YOUNGER", "KAKAK": "OLDER", "PACAR": "GIRLFRIEND",
	"SUAMI": "HUSBAND", "ISTRI": "WIFE", "TEMAN": "FRIEND", "MUSUH": "ENEMY",
	"HARI": "DAY", "MALAM": "NIGHT", "PAGI": "MORNING", "SIANG": "AFTERNOON",
	"SORE": "EVENING", "BESOK": "TOMORROW", "KEMARIN": "YESTERDAY", "MINGGU": "WEEK",
	"BULAN": "MONTH", "TAHUN": "YEAR", "JAM": "HOUR", "MENIT": "MINUTE",
	"DETIK": "SECOND",
	"SAMA": "WITH", "SAYANG": "LOVE", "RINDU": "MISS", "KENANG": "REMEMBER",
	"MANIS": "SWEET",
	"TENTANG": "ABOUT",
}

func TranslateIndonesian(text string) string {
	words := strings.Fields(strings.ToUpper(text))
	translated := make([]string, len(words))
	for i, word := range words {
		if eng, ok := indonesianDict[word]; ok {
			translated[i] = eng
		} else {
			translated[i] = word
		}
	}
	return strings.Join(translated, " ")
}

func IsIndonesian(text string) bool {
	words := strings.Fields(strings.ToUpper(text))
	for _, word := range words {
		if _, ok := indonesianDict[word]; ok {
			return true
		}
	}
	return false
}
