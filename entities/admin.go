package entities

import "time"

type AdminCreateUser struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type AdminUpdateUser struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type AdminUpdateUserResponse struct {
	Id    string        `json:"id"`
	Email string        `json:"email"`
	Phone string        `json:"phone"`
	Role  AdminRoleUser `json:"role"`
}

type AdminGetProfile struct {
	ID       string `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}

type AdminGetResponseProfile struct {
	ID       string `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
}

type AdminUpdateProfile struct {
	Fullname string `json:"fullname"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	UserId   string `json:"user_id"`
}

type AdminCreateUserResponse struct {
	Id    string        `json:"id"`
	Email string        `json:"email"`
	Phone string        `json:"phone"`
	Role  AdminRoleUser `json:"role"`
}

type AdminRoleUser struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AdminAssignRole struct {
	UserId string `json:"user_id"`
	RoleId string `json:"role_id"`
}

type AdminRevokeRole struct {
	UserId string `json:"user_id"`
}

type AdminListRole struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type AdminListRoleResponse struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type AdminListUser struct {
	Id               string    `json:"id"`
	Fullname         string    `json:"fullname"`
	Avatar           string    `json:"avatar"`
	Selfie           string    `json:"selfie"`
	PhotoKtp         string    `json:"photo_ktp"`
	NoKtp            string    `json:"no_ktp"`
	NoNpwp           string    `json:"no_npwp"`
	Position         string    `json:"position"`
	Gender           string    `json:"gender"`
	Sku              string    `json:"sku"`
	ProvinceName     string    `json:"province_name"`
	CityName         string    `json:"city_name"`
	DistrictName     string    `json:"district_name"`
	SubdistrictName  string    `json:"subdistrct_name"`
	StatusMarital    string    `json:"status_marital"`
	LastEducation    string    `json:"last_education"`
	AddressDetail    string    `json:"address_detail"`
	Occupation       string    `json:"occupation"`
	BeneficiaryName  string    `json:"beneficiary_name"`
	BeneficiaryPhone string    `json:"beneficiary_phone"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	Role             string    `json:"role"`
	Verify           bool      `json:"verify"`
	VerifyEmiten     bool      `json:"verify_emiten"`
	VerifyInvestor   bool      `json:"verify_investor"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AdminListUserResponse struct {
	Id               string             `json:"id"`
	Fullname         string             `json:"fullname"`
	Avatar           string             `json:"avatar"`
	Selfie           string             `json:"selfie"`
	PhotoKtp         string             `json:"photo_ktp"`
	NoKtp            string             `json:"no_ktp"`
	NoNpwp           string             `json:"no_npwp"`
	Jabatan          string             `json:"jabatan"`
	Gender           string             `json:"gender"`
	StatusMarital    string             `json:"status_marital"`
	LastEducation    string             `json:"last_education"`
	AddressDetail    string             `json:"address_detail"`
	Occupation       string             `json:"occupation"`
	ProvinceName     string             `json:"province_name"`
	CityName         string             `json:"city_name"`
	DistrictName     string             `json:"district_name"`
	SubdistrictName  string             `json:"subdistrict_name"`
	Email            string             `json:"email"`
	Phone            string             `json:"phone"`
	Sku              string             `json:"sku"`
	Role             string             `json:"role"`
	Verified         bool               `json:"verified"`
	VerifiedEmiten   bool               `json:"verified_emiten"`
	VerifiedInvestor bool               `json:"verified_investor"`
	NamaAhliWaris    string             `json:"nama_ahli_waris"`
	PhoneAhliWaris   string             `json:"phone_ahli_waris"`
	RekeningEfek     RekeningEfek       `json:"rekening_efek"`
	SuratKuasa       SuratKuasa         `json:"surat_kuasa"`
	Signature        AdminUserSignature `json:"signature"`
	Risk             AdminUserRisk      `json:"risk"`
	Emiten           AdminEmiten        `json:"emiten"`
	Investor         AdminInvestor      `json:"investor"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

type RekeningEfek struct {
	AccountName  string `json:"account_name"`
	AccountNo    string `json:"account_no"`
	AccountSubNo string `json:"account_sub_no"`
	AccountBank  string `json:"account_bank"`
}

type SuratKuasa struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type AdminDetailuser struct {
	UserId string `json:"user_id"`
}

type AdminListProject struct {
	Id                         string    `json:"id"`
	Title                      string    `json:"title"`
	Goal                       int64     `json:"goal"`
	Capital                    int64     `json:"capital"`
	Status                     string    `json:"status"`
	MinInvest                  int64     `json:"min_invest"`
	LoanTerm                   string    `json:"loan_term"`
	Site                       string    `json:"site"`
	UnitPrice                  int64     `json:"unit_price"`
	UnitTotal                  int64     `json:"unit_total"`
	CodeEffect                 string    `json:"code_effect"`
	NumberOfUnit               int64     `json:"number_of_init"`
	StartProject               string    `json:"start_project"`
	EndProject                 string    `json:"end_project"`
	Periode                    string    `json:"periode"`
	TypeOfProject              string    `json:"type_of_project"`
	NominalValue               int64     `json:"nominal_value"`
	TimePeriode                string    `json:"time_periode"`
	ProfitPercentage           string    `json:"profit_percentage"`
	TypeOfContractingAuthority string    `json:"type_of_contracting_authority"`
	RequiredFund               int64     `json:"required_fund"`
	Spk                        string    `json:"spk"`
	Loa                        string    `json:"loa"`
	Sku                        string    `json:"sku"`
	DocBankStatement           string    `json:"doc_bank_statement"`
	DocFinancialStatement      string    `json:"doc_financial_statement"`
	DocContract                string    `json:"doc_contract"`
	DocProspect                string    `json:"doc_prospect"`
	ContractingAuthority       string    `json:"contracting_authority"`
	ExpireDate                 string    `json:"expire_date"`
	InterestRate               string    `json:"interest_rate"`
	ProviderAddress            string    `json:"provider_address"`
	ProviderProvinceName       string    `json:"provider_province_name"`
	ProviderCityName           string    `json:"provider_city_name"`
	ProviderDistrictName       string    `json:"provider_district_name"`
	ProviderSubdistrictName    string    `json:"provider_subdistrict_name"`
	ProviderPostalCode         string    `json:"provider_postal_code"`
	PrincipalPaymentSchedule   string    `json:"principal_payment_schedule"`
	InterestPaymentSchedule    string    `json:"interest_payment_schedule"`
	UseOfFunds                 string    `json:"use_of_funds"`
	AmountSharesPerLot         int64     `json:"amount_shares_per_lot"`
	CollateralGuarantee        string    `json:"collateral_guarantee"`
	DescJob                    string    `json:"desc_job"`
	IsApbn                     bool      `json:"is_apbn"`
	IsApproved                 bool      `json:"is_approved"`
	UserId                     string    `json:"user_id"`
	UserEmail                  string    `json:"user_email"`
	UserName                   string    `json:"user_name"`
	UserPhone                  string    `json:"user_phone"`
	CreatedAt                  time.Time `json:"created_at"`
}

type AdminEmiten struct {
	Company AdminEmitenCompanyResponse `json:"company"`
}

type AdminEmitenCompany struct {
	Id                        string `json:"id"`
	Name                      string `json:"name"`
	Nib                       string `json:"nib"`
	NibPath                   string `json:"nib_path"`
	AktaPendirian             string `json:"akta_pendirian"`
	AktaPerubahanTerahkir     string `json:"akta_perubahan_terahkir"`
	AktaPerubahanTerahkirPath string `json:"akta_perubahan_terahkir_path"`
	SkPendirianPerusahaan     string `json:"sk_pendirian_perusahaan"`
	SkKumham                  string `json:"sk_kumham"`
	SkKumhamLast              string `json:"sk_kumham_last"`
	SkKumhamPath              string `json:"sk_kumham_path"`
	Npwp                      string `json:"npwp"`
	NpwpPath                  string `json:"npwp_path"`
	Site                      string `json:"site"`
	Email                     string `json:"email"`
	Phone                     string `json:"phone"`
	BankName                  string `json:"bank_name"`
	BankAccount               string `json:"bank_account"`
	BankOwnerCompany          string `json:"bank_owner_company"`
	Siup                      string `json:"siup"`
	Tdp                       string `json:"tdp"`
	RekeningKoran             string `json:"rekening_koran"`
	Est                       string `json:"est"`
	JenisPerusahaan           string `json:"jenis_perusahaan"`
	JenisUsaha                string `json:"jenis_usaha"`
	StatusKantor              string `json:"status_kantor"`
	TotalEmployees            string `json:"total_employees"`
	LaporanKeuangan           string `json:"laporan_keuangan"`
}

type AdminEmitenCompanyResponse struct {
	Id                        string                              `json:"id"`
	Name                      string                              `json:"name"`
	Nib                       string                              `json:"nib"`
	NibPath                   string                              `json:"nib_path"`
	AktaPendirian             string                              `json:"akta_pendirian"`
	AktaPerubahanTerahkir     string                              `json:"akta_perubahan_terahkir"`
	AktaPerubahanTerahkirPath string                              `json:"akta_perubahan_terahkir_path"`
	SkPendirianPerusahaan     string                              `json:"sk_pendirian_perusahaan"`
	SkKumham                  string                              `json:"sk_kumham"`
	SkKumhamTerahkir          string                              `json:"sk_kumham_terahkir"`
	SkKumhamPath              string                              `json:"sk_kumham_path"`
	Npwp                      string                              `json:"npwp"`
	NpwpPath                  string                              `json:"npwp_path"`
	Site                      string                              `json:"site"`
	Email                     string                              `json:"email"`
	Phone                     string                              `json:"phone"`
	BankName                  string                              `json:"bank_name"`
	BankAccount               string                              `json:"bank_account"`
	BankOwnerCompany          string                              `json:"bank_owner_company"`
	Siup                      string                              `json:"siup"`
	Tdp                       string                              `json:"tdp"`
	Est                       string                              `json:"est"`
	JenisPerusahaan           string                              `json:"jenis_perusahaan"`
	JenisUsaha                string                              `json:"jenis_usaha"`
	StatusKantor              string                              `json:"status_kantor"`
	TotalEmployees            string                              `json:"total_employees"`
	LaporanKeuangan           string                              `json:"laporan_keuangan"`
	RekeningKoran             string                              `json:"rekening_koran"`
	Address                   []AdminEmitenAddressCompany         `json:"address"`
	Positions                 []AdminEmitenPositionCompany        `json:"positions"`
	Projects                  []AdminEmitenProjectCompanyResponse `json:"projects"`
}

type AdminEmitenPositionCompany struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	Position string `json:"position"`
	Ktp      string `json:"ktp"`
	KtpPath  string `json:"ktp_path"`
	Npwp     string `json:"npwp"`
	NpwpPath string `json:"npwp_path"`
}

type AdminEmitenProjectCompany struct {
	Id                      string `json:"id"`
	Title                   string `json:"title"`
	TypeOfProject           string `json:"type_of_project"`
	JenisProject            string `json:"jenis_project"`
	JumlahMinimal           string `json:"jumlah_minimal"`
	JangkaWaktu             string `json:"jangka_waktu"`
	TingkatBunga            string `json:"tingkat_bunga"`
	StartProject            string `json:"start_project"`
	EndProject              string `json:"end_project"`
	JadwalPembayaranBunga   string `json:"jadwal_pembayaran_bunga"`
	JadwalPembayaranPokok   string `json:"jadwal_pembayaran_pokok"`
	DeskripsiPekerjaan      string `json:"deskripsi_pekerjaan"`
	CompanyProfile          string `json:"company_profile"`
	Website                 string `json:"website"`
	Spk                     string `json:"spk"`
	Loa                     string `json:"loa"`
	Sku                     string `json:"sku"`
	TenorPinjaman           string `json:"tenor_pinjaman"`
	IsApbn                  bool   `json:"is_apbn"`
	IsApproved              bool   `json:"is_approved"`
	Status                  string `json:"status"`
	ProviderAddress         string `json:"provider_address"`
	ProviderProvinceName    string `json:"provider_province_name"`
	ProviderCityName        string `json:"provider_city_name"`
	ProviderDistrictName    string `json:"provider_district_name"`
	ProviderSubdistrictName string `json:"provider_subdistrict_name"`
	ProviderPostalCode      string `json:"provider_postal_code"`
}

type AdminEmitenProjectCompanyResponse struct {
	Id                     string                            `json:"id"`
	Title                  string                            `json:"title"`
	JenisProject           string                            `json:"jenis_project"`
	JumlahMinimal          string                            `json:"jumlah_minimal"`
	JangkaWaktu            string                            `json:"jangka_waktu"`
	TingkatBunga           string                            `json:"tingkat_bunga"`
	JadwalPembayaranBunga  string                            `json:"jadwal_pembayaran_bunga"`
	JadwalPembayaranPokok  string                            `json:"jadwal_pembayaran_pokok"`
	DeskripsiPekerjaan     string                            `json:"deskripsi_pekerjaan"`
	CompanyProfile         string                            `json:"company_profile"`
	MulaiProject           string                            `json:"mulai_project"`
	SelesaiProject         string                            `json:"selesai_project"`
	PenggunaanData         []AdminProjectUseOfFunds          `json:"penggunaan_dana"`
	Sku                    string                            `json:"sku"`
	JaminanKolateral       []AdminProjectCollateralGuarantee `json:"jaminan_kolateral"`
	DocumentVerify         ProjectDocumentVerify             `json:"document_verify"`
	Kontrak                AdminProjectContract              `json:"kontrak"`
	TenorPinjaman          string                            `json:"tenor_pinjaman"`
	Website                string                            `json:"website"`
	IsApbn                 bool                              `json:"is_apbn"`
	IsApproved             bool                              `json:"is_approved"`
	Status                 string                            `json:"status"`
	AlamatPenyediaProject  string                            `json:"alamat_penyedia_project"`
	AlamatPenyediaProvinsi string                            `json:"alamat_penyedia_provinsi"`
	AlamatPenyediaKota     string                            `json:"alamat_penyedia_kota"`
	AlamatPenyediaDaerah   string                            `json:"alamat_penyedia_daerah"`
	AlamatPenyediaWilayah  string                            `json:"alamat_penyedia_wilayah"`
	AlamatPenyediaKodePos  string                            `json:"alamat_penyedia_kode_pos"`
	Medias                 []ProjectMediaPath                `json:"medias"`
}

type ProjectMediaPath struct {
	Path string `json:"path"`
}

type AdminInvestor struct {
	Ktp      AdminUserKtp     `json:"ktp"`
	Bank     AdminUserBank    `json:"bank"`
	Job      AdminUserJob     `json:"job"`
	SlipGaji AdminUserSlipPay `json:"slip_gaji"`
}

type AdminEmitenAddressCompany struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	Province    string `json:"province"`
	City        string `json:"city"`
	District    string `json:"district"`
	Subdistrict string `json:"subdistrict"`
	PostalCode  string `json:"postal_code"`
}

type AdminUserKtp struct {
	Name           string `json:"name"`
	Nik            string `json:"nik"`
	Path           string `json:"path"`
	PlaceDatebirth string `json:"place_datebirth"`
}

type AdminUserBank struct {
	No           string `json:"no"`
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	Branch       string `json:"branch"`
	RekKoranPath string `json:"rek_koran_path"`
}

type AdminUserSecurityAccount struct {
	AccountName  string `json:"account_name"`
	AccountNo    string `json:"account_no"`
	AccountSubNo string `json:"account_sub_no"`
	AccountBank  string `json:"account_bank"`
}

type AdminUserSignature struct {
	Path string `json:"path"`
}

type AdminUserRisk struct {
	Goal                  string `json:"goal"`
	Tolerance             string `json:"tolerance"`
	Experience            string `json:"experience"`
	PengetahuanPasarModal string `json:"pengetahuan_pasar_modal"`
}

type AdminUserSlipPay struct {
	Path string `json:"path"`
}

type AdminUserJob struct {
	CompanyName     string `json:"company_name"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
	CompanyAddress  string `json:"company_address"`
	MonthlyIncome   string `json:"monthly_income"`
	AnnualIncome    string `json:"annual_income"`
	Position        string `json:"position"`
	Npwp            string `json:"npwp"`
	NpwpPath        string `json:"npwp_path"`
}

type AdminListProjectResponse struct {
	Id                          string                            `json:"id"`
	Title                       string                            `json:"title"`
	Deskripsi                   string                            `json:"deskripsi"`
	Modal                       int64                             `json:"modal"`
	JumlahMinimal               int64                             `json:"jumlah_minimal"`
	JadwalPembayaranBunga       string                            `json:"jadwal_pembayaran_bunga"`
	JadwalPembayaranPokok       string                            `json:"jadwal_pembayaran_pokok"`
	PersentaseKeuntungan        string                            `json:"persentase_keuntungan"`
	JangkaWaktu                 string                            `json:"jangka_waktu"`
	TingkatBunga                string                            `json:"tingkat_bunga"`
	Spk                         string                            `json:"spk"`
	Loa                         string                            `json:"loa"`
	JumlahUnit                  int64                             `json:"jumlah_unit"`
	HargaPerlembar              int64                             `json:"harga_perlembar"`
	HargaPerlot                 int64                             `json:"harga_perlot"`
	JumlahLot                   int64                             `json:"jumlah_lot"`
	UnitTotal                   int64                             `json:"unit_total"`
	KodeEfek                    string                            `json:"kode_efek"`
	Sku                         string                            `json:"sku"`
	JenisProject                string                            `json:"jenis_project"`
	IsApbn                      bool                              `json:"is_apbn"`
	BuktiPembayaran             BuktiPembayaran                   `json:"bukti_pembayaran"`
	DocumentVerify              ProjectDocumentVerify             `json:"document_verify"`
	Kontrak                     AdminProjectContract              `json:"kontrak"`
	BatasAkhirPengerjaan        string                            `json:"batas_akhir_pengerjaan"`
	PenggunaanDana              []AdminProjectUseOfFunds          `json:"penggunaan_dana"`
	DanaYangDibutuhkan          int64                             `json:"dana_yang_dibutuhkan"`
	JaminanKolateral            []AdminProjectCollateralGuarantee `json:"jaminan_kolateral"`
	Website                     string                            `json:"website"`
	TenorPinjaman               string                            `json:"tenor_pinjaman"`
	IsApproved                  bool                              `json:"is_approved"`
	Status                      string                            `json:"status"`
	DocRekeningKoran            string                            `json:"doc_rekening_koran"`
	DocLaporanKeuangan          string                            `json:"doc_laporan_keuangan"`
	DocContract                 string                            `json:"doc_contract"`
	DocProspect                 string                            `json:"doc_prospect"`
	JenisInstansiPemberiProject string                            `json:"jenis_instansi_pemberi_project"`
	InstansiPemberiProject      string                            `json:"instansi_pemberi_project"`
	MulaiProject                string                            `json:"mulai_project"`
	SelesaiProject              string                            `json:"selesai_project"`
	AlamatPenyediaProject       string                            `json:"alamat_penyedia_project"`
	AlamatPenyediaProvinsi      string                            `json:"alamat_penyedia_provinsi"`
	AlamatPenyediaKota          string                            `json:"alamat_penyedia_kota"`
	AlamatPenyediaDaerah        string                            `json:"alamat_penyedia_daerah"`
	AlamatPenyediaWilayah       string                            `json:"alamat_penyedia_wilayah"`
	AlamatPenyediaKodePos       string                            `json:"alamat_penyedia_kode_pos"`
	Company                     AdminListCompany                  `json:"company"`
	Media                       []AdminListMedia                  `json:"media"`
	Location                    AdminProjectLocation              `json:"location"`
	User                        AdminListUserResponse             `json:"user"`
	CreatedAt                   time.Time                         `json:"created_at"`
}

type DocumentVerify struct {
	Skd                 string `json:"skd"`
	Cv                  string `json:"cv"`
	Rab                 string `json:"rab"`
	DokumenPerizinan    string `json:"dokumen_perizinan"`
	VideoProfileCompany string `json:"video_profile_company"`
	ProjectSummary      string `json:"project_summary"`
	ProjectPendapatan   string `json:"project_pendapatan"`
	TimelinePekerjaan   string `json:"timeline_pekerjaan"`
	LaporanPajakTahunan string `json:"laporan_pajak_tahunan"`
}

type DocumentMediaVerify struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type BuktiPembayaran struct {
	Path      string `json:"path"`
	IsApprove bool   `json:"is_approve"`
}

type AdminProjectLocation struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
	Lat  string `json:"lat"`
	Lng  string `json:"lng"`
}

type AdminProjectUseOfFunds struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type AdminProjectContract struct {
	Value string `json:"value"`
	Path  string `json:"path"`
}

type AdminProjectCollateralGuarantee struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type AdminListCompany struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Province    string `json:"province"`
	City        string `json:"city"`
	District    string `json:"district"`
	Subdistrict string `json:"subdistrict"`
}

type AdminListMedia struct {
	Id   int    `json:"id"`
	Path string `json:"path"`
}

type AdminListProjectUser struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type AdminVerifyUser struct {
	UserId string `json:"user_id"`
	Email  string `json:"email"`
}

type AdminUpdateProject struct {
	KodeEfek           string `json:"kode_efek"`
	JumlahUnit         string `json:"jumlah_unit"`
	HargaUnit          string `json:"harga_unit"`
	MinInvest          string `json:"min_invest"`
	ProfitPercentage   string `json:"profit_percentage"`
	AmountSharesPerLot string `json:"amount_shares_per_lot"`
	TotalUnit          string `json:"total_unit"`
	Tenor              string `json:"tenor"`
	Prospoectus        string `json:"prospectus"`
	ProjectId          string `json:"project_id"`
}

type AdminUpdateProjectResponse struct {
	ProjectId          string `json:"project_id"`
	KodeEfek           string `json:"kode_efek"`
	JumlahUnit         string `json:"jumlah_unit"`
	HargaUnit          string `json:"harga_unit"`
	MinInvest          string `json:"min_invest"`
	ProfitPercentage   string `json:"profit_percentage"`
	AmountSharesPerLot string `json:"amount_shares_per_lot"`
	TotalUnit          string `json:"total_unit"`
	Tenor              string `json:"tenor"`
	Prospoectus        string `json:"prospectus"`
}

type AdminVerifyProject struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}
