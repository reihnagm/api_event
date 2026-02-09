package entities

import "time"

type Register struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Fullname string `json:"fullname"`
	Role     string `json:"role"`
	Password string `json:"password"`
	UserId   string `json:"user_id"`
}

type RegisterAsEmiten struct {
	Fullname    string `json:"fullname"`
	PhotoSelfie string `json:"photo_selfie"`
	Jabatan     string `json:"jabatan"`
	PhotoKtp    string `json:"photo_ktp"`
	NoKtp       string `json:"no_ktp"`
	NoNpwp      string `json:"no_npwp"`
	SuratKuasa  string `json:"surat_kuasa"`
	Role        string `json:"role"`
	UserId      string `json:"user_id"`
}

type Bank struct {
	No     string `json:"no"`
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	Branch string `json:"branch"`
}

type Job struct {
	CompanyName    string `json:"company_name"`
	CompanyAddress string `json:"company_address"`
	MonthlyIncome  string `json:"monthly_income"`
	Position       string `json:"position"`
}

type Emiten struct {
	CompanyName         string   `json:"company_name"`
	CompanyData         string   `json:"company_data"`
	CompanyNib          string   `json:"company_nib"`
	DeedOfIncorporation string   `json:"deed_of_incorporation"`
	LatestAmendmentDeed string   `json:"latest_amendment_deed"`
	SkKumham            string   `json:"sk_kumham"`
	CompanyAddress      string   `json:"company_address"`
	CompanyNpwp         string   `json:"company_npwp"`
	TotalEmployees      string   `json:"total_employees"`
	CapitalStructure    string   `json:"capital_structure"`
	FinancialStatement  string   `json:"financial_statement"`
	CommisionerName     string   `json:"commisioner_name"`
	CommisionerPosition string   `json:"commisioner_position"`
	CommisionerKtp      string   `json:"commisioner_ktp"`
	CommisionerNpwp     string   `json:"commisioner_npwp"`
	DirectorName        string   `json:"director_name"`
	DirectorPosition    string   `json:"director_position"`
	DirectorKtp         string   `json:"director_ktp"`
	DirectorNpwp        string   `json:"director_npwp"`
	InfoBond            InfoBond `json:"info_bond"`
}

type InfoBond struct {
	Title                    string   `json:"title"`
	Img                      string   `json:"img"`
	Doc                      string   `json:"doc"`
	Location                 Location `json:"location"`
	TypeOfProject            string   `json:"type_of_project"`
	NominalValue             string   `json:"nominal_value"`
	TimePeriode              string   `json:"time_periode"`
	InterestRate             string   `json:"interest_rate"`
	InterestPaymentSchedule  string   `json:"interest_payment_schedule"`
	PrincipalPaymentSchedule string   `json:"principal_payment_schedule"`
	UseOfFunds               string   `json:"use_of_funds"`
	CollateralGuarantee      string   `json:"collateral_guarantee"`
	DescJob                  string   `json:"desc_job"`
	IsApbn                   bool     `json:"is_apbn"`
}

type Location struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Lat  string `json:"lat"`
	Lng  string `json:"lng"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Logout struct {
	Email string `json:"email"`
}

type AdminLogin struct {
	Val      string `json:"val"`
	Password string `json:"password"`
}

type UserOtp struct {
	Id        string    `json:"uid"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Enabled   int       `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type ResendOtp struct {
	Val string `json:"val"`
}

type VerifyOtp struct {
	Val string `json:"val"`
	Otp string `json:"otp"`
}

type LoginScan struct {
	Id       string `json:"id"`
	Enabled  bool   `json:"enabled"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Verify   bool   `json:"verify"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Id      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Verify  bool   `json:"verify"`
	Token   string `json:"token"`
}

type LogoutResponse struct {
	Email string `json:"email"`
}

type RegisterResponse struct {
	Id      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Verify  bool   `json:"verify"`
	Token   string `json:"token"`
}

type VerifyOtpResponse struct {
	Id      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	Verify  bool   `json:"verify"`
	Token   string `json:"token"`
}

type CheckRole struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AssignRole struct {
	Ktp                       InvestorKtp                `json:"ktp"`
	Bank                      InvestorBank               `json:"bank"`
	Job                       InvestorJob                `json:"job"`
	Risk                      InvestorRisk               `json:"risk"`
	Goal                      string                     `json:"goal"`
	NamaAhliWaris             string                     `json:"nama_ahli_waris"`
	PhoneAhliWaris            string                     `json:"phone_ahli_waris"`
	Siup                      string                     `json:"siup"`
	Tdp                       string                     `json:"tdp"`
	Site                      string                     `json:"site"`
	Tolerance                 string                     `json:"tolerance"`
	Experience                string                     `json:"experience"`
	CapitalMarketKnowledge    string                     `json:"capital_market_knowledge"`
	Occupation                string                     `json:"occupation"`
	CompanyName               string                     `json:"company_name"`
	CompanyNib                string                     `json:"company_nib"`
	CompanyNibPath            string                     `json:"company_nib_path"`
	AktaPendirian             string                     `json:"akta_pendirian"`
	AktaPerubahanTerahkir     string                     `json:"akta_perubahan_terahkir"`
	AktaPendirianPath         string                     `json:"akta_pendirian_path"`
	AktaPerubahanTerahkirPath string                     `json:"akta_perubahan_terahkir_path"`
	SkKumham                  string                     `json:"sk_kumham"`
	SkKumhamTerahkir          string                     `json:"sk_kumham_terahkir"`
	SkKumhamPath              string                     `json:"sk_kumham_path"`
	Npwp                      string                     `json:"npwp"`
	NpwpPath                  string                     `json:"npwp_path"`
	TotalEmployees            string                     `json:"total_employees"`
	LaporanKeuanganPath       string                     `json:"laporan_keuangan_path"`
	RekeningKoranPath         string                     `json:"rekening_koran_path"`
	Directors                 []AssignRolePosition       `json:"directors"`
	Komisaris                 []AssignRolePosition       `json:"komisaris"`
	Avatar                    string                     `json:"avatar"`
	Role                      string                     `json:"role"`
	Didirikan                 string                     `json:"didirikan"`
	Email                     string                     `json:"email"`
	Phone                     string                     `json:"phone"`
	Gender                    string                     `json:"gender"`
	BankName                  string                     `json:"bank_name"`
	BankAccount               string                     `json:"bank_account"`
	BankOwner                 string                     `json:"bank_owner"`
	StatusMarital             string                     `json:"status_marital"`
	LastEducation             string                     `json:"last_education"`
	ProvinceName              string                     `json:"province_name"`
	CityName                  string                     `json:"city_name"`
	DistrictName              string                     `json:"district_name"`
	SubdistrictName           string                     `json:"subdistrict_name"`
	PostalCode                string                     `json:"postal_code"`
	AddressDetail             string                     `json:"address_detail"`
	Address                   []UpdateRoleCompanyAddress `json:"address"`
	Project                   AssignRoleProject          `json:"project"`
	SignaturePath             string                     `json:"signature_path"`
	UserId                    string                     `json:"user_id"`
	SkPendirianPerusahaan     string                     `json:"sk_pendirian_perusahaan"`
	SlipGaji                  string                     `json:"slip_gaji"`
	NamaRekeningEfek          string                     `json:"nama_rekening_efek"`
	NomorRekeningEfek         string                     `json:"nomor_rekening_efek"`
	NomorSubRekeningEfek      string                     `json:"nomor_sub_rekening_efek"`
	BankRekeningEfek          string                     `json:"bank_rekening_efek"`
	JenisPerusahaan           string                     `json:"jenis_perusahaan"`
	JenisUsaha                string                     `json:"jenis_usaha"`
	StatusKantor              string                     `json:"status_kantor"`
}

type InvestorRisk struct {
	Goal                  string `json:"goal"`
	Tolerance             string `json:"tolerance"`
	Experience            string `json:"experience"`
	PengetahuanPasarModal string `json:"pengetahuan_pasar_modal"`
}

type InvestorKtp struct {
	Name           string `json:"name"`
	PlaceDatebirth string `json:"place_datebirth"`
	Nik            string `json:"nik"`
	NikPath        string `json:"nik_path"`
}

type InvestorJob struct {
	Company         string `json:"company"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	PostalCode      string `json:"postal_code"`
	Address         string `json:"address"`
	MonthlyIncome   string `json:"monthly_income"`
	AnnualIncome    string `json:"annual_income"`
	Position        string `json:"position"`
	Npwp            string `json:"npwp"`
	NpwpPath        string `json:"npwp_path"`
}

type InvestorBank struct {
	No           string `json:"no"`
	Name         string `json:"name"`
	Branch       string `json:"branch"`
	Owner        string `json:"owner"`
	RekKoranPath string `json:"rek_koran_path"`
}

type AssignRolePosition struct {
	Title    string `json:"title"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Ktp      string `json:"ktp"`
	KtpPath  string `json:"ktp_path"`
	Npwp     string `json:"npwp"`
	NpwpPath string `json:"npwp_path"`
}

type AssignRoleProject struct {
	Title                 string                          `json:"title"`
	Goal                  string                          `json:"goal"`
	JenisProjek           string                          `json:"jenis_projek"`
	JumlahMinimal         string                          `json:"jumlah_minimal"`
	JangkaWaktu           string                          `json:"jangka_waktu"`
	TingkatBunga          string                          `json:"tingkat_bunga"`
	Capital               string                          `json:"capital"`
	MinInvest             string                          `json:"min_invest"`
	UnitPrice             string                          `json:"unit_price"`
	UnitTotal             string                          `json:"unit_total"`
	NumberOfUnit          string                          `json:"number_of_unit"`
	Periode               string                          `json:"periode"`
	JadwalPembayaranBunga string                          `json:"jadwal_pembayaran_bunga"`
	JadwalPembayaranPokok string                          `json:"jadwal_pembayaran_pokok"`
	PenggunaanDana        []AssignProjectPenggunaanDana   `json:"penggunaan_dana"`
	CompanyProfile        string                          `json:"company_profile"`
	JaminanKolateral      []AssignProjectJaminanKolateral `json:"jaminan_kolateral"`
	DeskripsiPekerjaan    string                          `json:"deskripsi_pekerjaan"`
	ProjectMediaPath      []string                        `json:"project_media_path"`
	NoContractValue       string                          `json:"no_contract_value"`
	NoContractPath        string                          `json:"no_contract_path"`
	Doc                   AssignProjectDoc                `json:"doc"`
	Location              AssignProjectLocation           `json:"location"`
	IsApbn                bool                            `json:"is_apbn"`
}

type AssignProjectPenggunaanDana struct {
	Name string `json:"name"`
}

type AssignProjectJaminanKolateral struct {
	Name string `json:"name"`
}

type AssignProjectDoc struct {
	Id   string `json:"id"`
	Path string `json:"path"`
}

type AssignProjectLocation struct {
	Name string `json:"name"`
	Url  string `json:"url"`
	Lat  string `json:"lat"`
	Lng  string `json:"lng"`
}

type UpdateRoleCompanyAddress struct {
	Name            string `json:"name"`
	Detail          string `json:"detail"`
	PostalCode      string `json:"postal_code"`
	ProvinceId      string `json:"province_id"`
	ProvinceName    string `json:"province_name"`
	CityId          string `json:"city_id"`
	CityName        string `json:"city_name"`
	DistrictId      string `json:"district_id"`
	DistrictName    string `json:"district_name"`
	SubdistrictId   string `json:"subdistrict_id"`
	SubdistrictName string `json:"subdistrict_name"`
}

type UpdateRoleResponse struct {
	CompanyName           string `json:"company_name"`
	CompanyNib            string `json:"company_nib"`
	CompanyNibPath        string `json:"company_nib_path"`
	AktaPendirian         string `json:"akta_pendirian"`
	AktaPerubahanTerahkir string `json:"akta_perubahan_terahkir"`
	SkKumham              string `json:"sk_kumham"`
	SkKumhamPath          string `json:"sk_kumham_path"`
	NpwpPath              string `json:"npwp_path"`
	TotalEmployees        string `json:"total_employees"`
	LaporanKeuangan       string `json:"laporan_keuangan"`
}
