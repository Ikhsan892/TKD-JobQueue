package dto

type ReportInventoryResponse struct {
	KodeKlasifikasi         string `json:"kode_klasifikasi"`
	JudulArsip              string `json:"judul_arsip"`
	FrekuensiPenambahan     string `json:"frekuensi_penambahan"`
	TahunDari               string `json:"tahun_dari"`
	TahunSampai             string `json:"tahun_sampai"`
	MediaSimpan             string `json:"media_simpan"`
	SaranaSimpan            string `json:"sarana_simpan"`
	IsiJenisDokumen         string `json:"isi_jenis_dokumen"`
	BentukDokumen           string `json:"bentuk_dokumen"`
	UkuranFisikDimensiArsip string `json:"ukuran_fisik_dimensi_arsip"`
	TingkatKeaslian         string `json:"tingkat_keaslian"`
	DiisiOleh               string `json:"diisi_oleh"`
	Profile                 string `json:"profile"`
}

type ReportInventory struct {
	ReportVolume
}
