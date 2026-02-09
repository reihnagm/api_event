package entities

import "time"

// ProfileScan hanya untuk hasil scan dari `profiles`
type ProfileScan struct {
	Id               string `json:"id"`
	Fullname         string `json:"fullname"`
	Role             string `json:"role"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	Avatar           string `json:"avatar"`
	Selfie           string `json:"selfie"`
	Gender           string `json:"gender"`
	LastEducation    string `json:"last_education"`
	StatusMarital    string `json:"status_marital"`
	Occupation       string `json:"occupation"`
	Position         string `json:"position"`
	ProvinceName     string `json:"province_name"`
	CityName         string `json:"city_name"`
	DistrictName     string `json:"district_name"`
	SubdistrictName  string `json:"subdistrict_name"`
	BeneficiaryName  string `json:"beneficiary_name"`
	BeneficiaryPhone string `json:"beneficiary_phone"`
	PostalCode       string `json:"postal_code"`
	AddressDetail    string `json:"address_detail"`
	PhotoKtp         string `json:"photo_ktp"`
	NoKtp            string `json:"no_ktp"`
	NoNpwp           string `json:"no_npwp"`
	VerifyEmiten     bool   `json:"verify_emiten"`
	VerifyInvestor   bool   `json:"verify_investor"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// Hasil scan dari `companies`
type ProfileEmiten struct {
	CompanyName          string `json:"company_name"`
	CompanyData          string `json:"company_data"`
	CompanyNib           string `json:"company_nib"`
	CompanyAddress       string `json:"company_address"`
	CompanyNpwp          string `json:"company_npwp"`
	CapitalStructure     string `json:"capital_structure"`
	FinancialStatement   string `json:"financial_statement"`
	CommissionerName     string `json:"commissioner_name"`
	CommissionerPosition string `json:"commissioner_position"`
	CommissionerKtp      string `json:"commissioner_ktp"`
	CommissionerNpwp     string `json:"commissioner_npwp"`
	DirectorName         string `json:"director_name"`
	DirectorPosition     string `json:"director_position"`
	DirectorKtp          string `json:"director_ktp"`
	DirectorNpwp         string `json:"director_npwp"`
	TotalEmployees       string `json:"total_employees"`
	DeedOfIncorporation  string `json:"deed_of_incorporation"`
	LatestAmendmentDeed  string `json:"latest_amendment_deed"`
	SkKumham             string `json:"sk_kumham"`
}

type ProfileInvestor struct {
	Gender        string      `json:"gender"`
	LastEdu       string      `json:"last_edu"`
	StatusMarital string      `json:"status_marital"`
	Ktp           string      `json:"ktp"`
	AddressKtp    string      `json:"address_ktp"`
	Bank          ProfileBank `json:"bank"`
	Job           ProfileJob  `json:"job"`
}

type ProfileBank struct {
	No     string `json:"no"`
	Name   string `json:"name"`
	Owner  string `json:"owner"`
	Branch string `json:"branch"`
}

type ProfileJob struct {
	CompanyName    string `json:"company_name"`
	CompanyAddress string `json:"company_address"`
	MonthlyIncome  string `json:"monthly_income"`
	Position       string `json:"position"`
}

type ProfileResponse struct {
	Profile
}

type ActiveProj struct {
	Id         string
	Title      string
	StatusName string
}

type ActiveProjFlag struct {
	Id         string
	Title      string
	StatusName string
}

type Profile struct {
	Id                     string                 `json:"id"`
	Role                   string                 `json:"role"`
	Fullname               string                 `json:"fullname"`
	Avatar                 string                 `json:"avatar"`
	Selfie                 string                 `json:"selfie"`
	PhotoKtp               string                 `json:"photo_ktp"`
	NoKtp                  string                 `json:"no_ktp"`
	Npwp                   string                 `json:"npwp"`
	Phone                  string                 `json:"phone"`
	Email                  string                 `json:"email"`
	LastEducation          string                 `json:"last_education"`
	Gender                 string                 `json:"gender"`
	StatusMarital          string                 `json:"status_marital"`
	AddressDetail          string                 `json:"address_detail"`
	Occupation             string                 `json:"occupation"`
	Position               string                 `json:"position"`
	ProvinceName           string                 `json:"province_name"`
	CityName               string                 `json:"city_name"`
	DistrictName           string                 `json:"district_name"`
	SubdistrictName        string                 `json:"subdistrict_name"`
	PostalCode             string                 `json:"postal_code"`
	SlipGaji               string                 `json:"slip_gaji"`
	ProfileSecurityAccount ProfileSecurityAccount `json:"profile_security_account"`
	Doc                    ProfileDoc             `json:"doc"`
	Investor               ProfileUserInvestor    `json:"investor"`
	Company                ProfileUserCompany     `json:"company"`
	RekEfek                bool                   `json:"rek_efek"`
	VerifyEmiten           bool                   `json:"verify_emiten"`
	VerifyInvestor         bool                   `json:"verify_investor"`
	NamaAhliWaris          string                 `json:"nama_ahli_waris"`
	PhoneAhliWaris         string                 `json:"phone_ahli_waris"`
	CanCreateProject       bool                   `json:"can_create_project"`
	CreatedAt              string                 `json:"created_at"`
	UpdatedAt              string                 `json:"updated_at"`
}

type ProfileDoc struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type ProfileUserInvestor struct {
	Bank ProfileUserInvestorBankScan `json:"bank"`
	Ktp  ProfileUserInvestorKtpScan  `json:"ktp"`
	Job  ProfileUserInvestorJobScan  `json:"job"`
	Risk ProfileuserInvestorRiskScan `json:"risk"`
}

type ProfileuserInvestorRiskScan struct {
	Goal                   string `json:"goal"`
	Tolerance              string `json:"tolerance"`
	Experience             string `json:"experience"`
	CapitalMarketKnowledge string `json:"capital_market_knowledge"`
}

type ProfileAdditionalDocScan struct {
	Path   string `json:"path"`
	Type   string `json:"type"`
	UserId string `json:"user_id"`
}

type ProfileUserInvestorBankScan struct {
	No           string    `json:"no"`
	BankName     string    `json:"bank_name"`
	BankOwner    string    `json:"bank_owner"`
	BankBranch   string    `json:"bank_branch"`
	RekKoranPath string    `json:"rek_koran_path"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProfileUserInvestorKtpScan struct {
	Name           string    `json:"name"`
	Nik            string    `json:"nik"`
	PlaceDatebirth string    `json:"place_datebirth"`
	Path           string    `json:"path"`
	CreatedAt      time.Time `json:"created_at"`
}

type ProfileSlipPay struct {
	Path string `json:"path"`
}

type ProfileSecurityAccount struct {
	AccountName  string `json:"account_name"`
	AccountNo    string `json:"account"`
	AccountSubNo string `json:"account_sub_no"`
	AccountBank  string `json:"account_bank"`
}

type ProfileUserInvestorJobScan struct {
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	PostalCode      string `json:"postal_code"`
	CompanyName     string `json:"company_name"`
	CompanyAddress  string `json:"company_address"`
	MonthlyIncome   string `json:"monthly_income"`
	AnnualIncome    string `json:"annual_income"`
	NpwpPath        string `json:"npwp_path"`
	Npwp            string `json:"npwp"`
	Position        string `json:"position"`
}

type ProfileUserCompanyScan struct {
	Id                      string `json:"id"`
	CompanyName             string `json:"company_name"`
	CompanyNib              string `json:"company_nib"`
	CompanyNibPath          string `json:"company_nib_path"`
	DeedOfIncorporation     string `json:"deed_of_incorporation"`
	LatestAmendmentDeed     string `json:"latest_amendment_deed"`
	LatestAmendmentDeedPath string `json:"latest_amendment_deed_path"`
	Siup                    string `json:"siup"`
	Tdp                     string `json:"tdp"`
	Site                    string `json:"site"`
	Email                   string `json:"email"`
	Npwp                    string `json:"npwp"`
	BankName                string `json:"bank_name"`
	BankAccount             string `json:"bank_account"`
	BankOwnerCompany        string `json:"bank_owner_company"`
	JenisPerusahaan         string `json:"jenis_perusahaan"`
	SkPendirianPerusahaan   string `json:"sk_pendirian_perusahaan"`
	SkKumham                string `json:"sk_kumham"`
	SkKumhamLast            string `json:"sk_kumham_latest"`
	SkKumhamPath            string `json:"sk_kumham_path"`
	NpwpPath                string `json:"npwp_path"`
	Phone                   string `json:"phone"`
	TotalEmployees          string `json:"total_employees"`
	FinancialStatement      string `json:"financial_statement"`
	BankStatement           string `json:"bank_statement"`
}

type ProfileUserProjectScan struct {
	Id                       string `json:"id"`
	Title                    string `json:"title"`
	TypeOfProject            string `json:"type_of_project"`
	NominalValue             int    `json:"nominal_value"`
	TimePeriode              string `json:"time_periode"`
	InterestRate             string `json:"interest_rate"`
	InterestPaymentSchedule  string `json:"interest_payment_schedule"`
	PrincipalPaymentSchedule string `json:"principal_payment_schedule"`
	DescJob                  string `json:"desc_job"`
	CompanyProfile           string `json:"company_profile"`
	Spk                      string `json:"spk"`
	Loa                      string `json:"loa"`
	StartProject             string `json:"start_project"`
	EndProject               string `json:"end_project"`
	Status                   string `json:"status"`
	IsApbn                   bool   `json:"is_apbn"`
}

type ProfileUserProjectContractScan struct {
	Value string `json:"value"`
	Path  string `json:"path"`
}

type ProfileUserCompany struct {
	Id                        string                      `json:"id"`
	Name                      string                      `json:"name"`
	Nib                       string                      `json:"nib"`
	NibPath                   string                      `json:"nib_path"`
	AktaPendirian             string                      `json:"akta_pendirian"`
	AktaPerubahanTerahkir     string                      `json:"akta_perubahan_terahkir"`
	AktaPerubahanTerahkirPath string                      `json:"akta_perubahan_terahkir_path"`
	SkKumham                  string                      `json:"sk_kumham"`
	SkKumhamTerahkir          string                      `json:"sk_kumham_terahkir"`
	SkPendirianPerusahaan     string                      `json:"sk_pendirian_perusahaan"`
	SkKumhamPath              string                      `json:"sk_kumham_path"`
	NpwpPath                  string                      `json:"npwp_path"`
	Siup                      string                      `json:"siup"`
	Tdp                       string                      `json:"tdp"`
	Site                      string                      `json:"site"`
	Email                     string                      `json:"email"`
	Npwp                      string                      `json:"npwp"`
	Phone                     string                      `json:"phone"`
	Bank                      ProfileBankCompany          `json:"bank"`
	JenisPerusahaan           string                      `json:"jenis_perusahaan"`
	TotalEmployees            string                      `json:"total_employees"`
	LaporanKeuanganPath       string                      `json:"laporan_keuangan_path"`
	RekeningKoran             string                      `json:"rekening_koran"`
	Address                   []ProfileUserAddressCompany `json:"address"`
	Directors                 []ProfileUserDirector       `json:"directors"`
	Komisaris                 []ProfileUserKomisaris      `json:"komisaris"`
	Projects                  []ProfileUserProject        `json:"projects"`
}

type ProfileBankCompany struct {
	Name  string `json:"name"`
	No    string `json:"no"`
	Owner string `json:"owner"`
}

type ProfileUserAddressCompany struct {
	Name            string `json:"name"`
	Detail          string `json:"detail"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	PostalCode      string `json:"postal_code"`
}

type ProfileUserDirector struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Ktp      string `json:"ktp"`
	KtpPath  string `json:"ktp_path"`
	Npwp     string `json:"npwp"`
	NpwpPath string `json:"npwp_path"`
}

type ProfileUserKomisaris struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Ktp      string `json:"ktp"`
	KtpPath  string `json:"ktp_path"`
	Npwp     string `json:"npwp"`
	NpwpPath string `json:"npwp_path"`
}

type ProfileCheckSecAccount struct {
}

type ProfileUserProject struct {
	Id                    string                               `json:"id"`
	Title                 string                               `json:"title"`
	Deskripsi             string                               `json:"deskripsi"`
	JenisProjek           string                               `json:"jenis_projek"`
	JumlahMinimal         int                                  `json:"jumlah_minimal"`
	JangkaWaktu           string                               `json:"jangka_waktu"`
	TingkatBunga          string                               `json:"tingkat_bunga"`
	CompanyProfile        string                               `json:"company_profile"`
	Spk                   string                               `json:"spk"`
	Loa                   string                               `json:"loa"`
	MulaiProject          string                               `json:"mulai_project"`
	SelesaiProject        string                               `json:"selesai_project"`
	JadwalPembayaranBunga string                               `json:"jadwal_pembayaran_bunga"`
	JadwalPembayaranPokok string                               `json:"jadwal_pembayaran_pokok"`
	Media                 []ProfileUserProjectMedia            `json:"media"`
	JaminanKolateral      []ProfileUserProjectJaminanKolateral `json:"jaminan_kolateral"`
	PenggunaanDana        []ProfileUserProjectPenggunaanData   `json:"penggunaan_dana"`
	NilaiKontrakPath      string                               `json:"nilai_kontrak_path"`
	NilaiKontrak          string                               `json:"nilai_kontrak"`
	Status                string                               `json:"status"`
	IsApbn                bool                                 `json:"is_apbn"`
}

type ProfileUserProjectMedia struct {
	Id   int    `json:"id"`
	Path string `json:"path"`
}

type ProfileUserProjectJaminanKolateral struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ProfileUserProjectPenggunaanData struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type GetProfile struct {
	UserId string `json:"user_id"`
}

type UpdateProfile struct {
	UserId           string `json:"user_id"`
	Fullname         string `json:"fullname"`
	Avatar           string `json:"avatar"`
	Selfie           string `json:"selfie"`
	Gender           string `json:"gender"`
	StatusMarital    string `json:"status_marital"`
	LastEducation    string `json:"last_education"`
	ProvinceName     string `json:"province_name"`
	CityName         string `json:"city_name"`
	DistrictName     string `json:"district_name"`
	SubdistrictName  string `json:"subdistrict_name"`
	PostalCode       string `json:"postal_code"`
	AddressDetail    string `json:"address_detail"`
	PhotoKtp         string `json:"photo_ktp"`
	NoKtp            string `json:"no_ktp"`
	Position         string `json:"position"`
	Occupation       string `json:"occupation"`
	BeneficiaryName  string `json:"beneficiary_name"`
	BeneficiaryPhone string `json:"beneficiary_phone"`
}
