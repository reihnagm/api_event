package entities

type DocumentVerifyProject struct {
	Skd                     string               `json:"skd"`
	DokumenPerizinanLainnya string               `json:"dokumen_perizinan_lainnya"`
	Cv                      string               `json:"cv"`
	VideoProfilPerusahaan   string               `json:"video_profil_perusahaan"`
	ProjectSummary          string               `json:"project_summary"`
	ProjectPendapatan       string               `json:"project_pendapatan"`
	TimelinePekerjaan       string               `json:"timeline_pekerjaan"`
	FotoKegiatanUsaha       []FotoKegiatanUsaha  `json:"foto_kegiatan_usaha"`
	FotoKaryawanKantor      []FotoKaryawanKantor `json:"foto_karyawan_kantor"`
	LaporanPajakTahunan     string               `json:"laporan_pajak_tahunan"`
	DaftarPekerjaan         string               `json:"daftar_pekerjaan"`
	DaftarSupplier          string               `json:"daftar_supplier"`
	DaftarPiutang           string               `json:"daftar_piutang"`
	CashflowProject         string               `json:"cashflow_project"`
	Rab                     string               `json:"rab"`
	InboxId                 string               `json:"inbox_id"`
	ProjectId               string               `json:"project_id"`
}

type FotoKegiatanUsaha struct {
	Path string `json:"path"`
}

type FotoKaryawanKantor struct {
	Path string `json:"path"`
}

type ContractLetterPaymentUpload struct {
	ProjectPaymentId string `json:"project_payment_id"`
	Path             string `json:"path"`
}

type DocumentMediaVerifyProject struct {
	DocumentVerifyProjectId string `json:"document_verify_project_id"`
	Path                    string `json:"path"`
	Type                    string `json:"type"`
}

type DocumentTransactionPayment struct {
	Path      string `json:"path"`
	InboxId   string `json:"inbox_id"`
	ProjectId string `json:"project_id"`
}

type ValidationError struct {
	Missing []string `json:"missing"`
}

type DocumentUpdate struct {
	CompanyId string     `json:"company_id"`
	ProjectId string     `json:"project_id"`
	InboxId   string     `json:"inbox_id"`
	UserId    string     `json:"user_id"`
	Val       string     `json:"val"`
	ValArray  []ValArray `json:"val_array"`
	Type      string     `json:"type"`
}

type ValArray struct {
	Id   string `json:"id"`
	Val  string `json:"val"`
	Type string `json:"type"`
}

type UpdateValUser struct {
	UserId string `json:"user_id"`
	Val    string `json:"val"`
	Type   string `json:"type"`
}
