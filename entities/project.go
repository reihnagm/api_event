package entities

import "time"

type ProjectListScan struct {
	Id                       string    `json:"id"`
	Title                    string    `json:"title"`
	Goal                     int64     `json:"goal"`
	TargetAmount             int64     `json:"target_amount"`
	UserPaid                 int64     `json:"user_paid"`
	InvestorsPaid            int64     `json:"investors_paid"`
	Capital                  int64     `json:"capital"`
	CodeEffect               string    `json:"code_effect"`
	CanRefund                bool      `json:"can_refund"`
	MinInvest                int64     `json:"min_invest"`
	UnitPrice                int64     `json:"unit_price"`
	UnitTotal                int64     `json:"unit_total"`
	NumberOfUnit             int64     `json:"number_of_unit"`
	ProfitPercentage         string    `json:"profit_percentage"`
	DocProspect              string    `json:"doc_prospect"`
	LoanTerm                 string    `json:"loan_term"`
	Periode                  string    `json:"periode"`
	AmountSharesPerLot       int64     `json:"amount_shares_per_lot"`
	TypeOfProject            string    `json:"type_of_project"`
	NominalValue             string    `json:"nominal_value"`
	StartProject             string    `json:"start_project"`
	EndProject               string    `json:"end_project"`
	TimePeriode              string    `json:"time_periode"`
	InterestRate             string    `json:"interest_rate"`
	UseOfFunds               string    `json:"use_of_funds"`
	InterestPaymentSchedule  string    `json:"interest_payment_schedule"`
	PrincipalPaymentSchedule string    `json:"principal_payment_schedule"`
	CollateralGuarantee      string    `json:"collateral_guarantee"`
	DescJob                  string    `json:"desc_job"`
	IsApbn                   bool      `json:"is_apbn"`
	IsApproved               bool      `json:"is_approved"`
	ProviderAddress          string    `json:"provider_address"`
	ProviderProvinceName     string    `json:"provider_province_name"`
	ProviderCityName         string    `json:"provider_city_name"`
	ProviderDistrictName     string    `json:"provider_district_name"`
	ProviderSubdistrictName  string    `json:"provider_subdistrict_name"`
	ProviderPostalCode       int64     `json:"provider_postal_code"`
	UserId                   string    `json:"user_id"`
	Status                   string    `json:"status"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type ProjectTypeListScan struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ProjectAuthorityTypeListScan struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ProjectCostOfFundTemplateListScan struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Percentage   float32   `json:"percentage"`
	FixedAmount  int       `json:"fixed_amount"`
	PaymentSplit string    `json:"payment_split"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProjectCostOfFundTemplateWithoutListScan struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Percentage  float32   `json:"percentage"`
	FixedAmount int       `json:"fixed_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectCostOfFundTemplateDetailScan struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Percentage   float32   `json:"percentage"`
	FixedAmount  int       `json:"fixed_amount"`
	PaymentSplit string    `json:"payment_split"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProjectCostOfFundTemplateWithoutDetailScan struct {
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Percentage  float32   `json:"percentage"`
	FixedAmount int       `json:"fixed_amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectCostOfFundTemplateStore struct {
	Name         string  `json:"name"`
	Percentage   float32 `json:"percentage"`
	FixedAmount  int     `json:"fixed_amount"`
	PaymentSplit string  `json:"payment_split"`
}

type ProjectCostOfFundTemplateWithoutStore struct {
	Name        string  `json:"name"`
	Percentage  float32 `json:"percentage"`
	FixedAmount int     `json:"fixed_amount"`
}

type ProjectCostOfFundTemplateUpdate struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Percentage   float32 `json:"percentage"`
	FixedAmount  int     `json:"fixed_amount"`
	PaymentSplit string  `json:"payment_split"`
}

type ProjectCostOfFundTemplateWithoutUpdate struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Percentage  float32 `json:"percentage"`
	Description string  `json:"description"`
	FixedAmount int     `json:"fixed_amount"`
}

type ProjectListResponse struct {
	Id                       string                       `json:"id"`
	Title                    string                       `json:"title"`
	Goal                     int64                        `json:"goal"`
	TargetAmount             int64                        `json:"target_amount"`
	UserPaidAmount           int64                        `json:"user_paid_amount"`
	InvestorPaid             int64                        `json:"investor_paid"`
	Medias                   []ProjectMedia               `json:"medias"`
	Location                 ProjectLocation              `json:"location"`
	Doc                      ProjectDoc                   `json:"doc"`
	Capital                  int64                        `json:"capital"`
	KodeEfek                 string                       `json:"kode_efek"`
	LoanTerm                 string                       `json:"loan_term"`
	Roi                      string                       `json:"roi"`
	MinInvest                int64                        `json:"min_invest"`
	HargaUnit                int64                        `json:"unit_price"`
	UnitTotal                int64                        `json:"unit_total"`
	JumlahLot                int64                        `json:"jumlah_lot"`
	JumlahUnit               int64                        `json:"jumlah_unit"`
	StokLot                  int64                        `json:"stok_lot"`
	Periode                  string                       `json:"periode"`
	DocProspect              string                       `json:"doc_prospect"`
	TypeOfProject            string                       `json:"type_of_project"`
	NominalValue             string                       `json:"nominal_value"`
	TimePeriode              string                       `json:"time_periode"`
	InterestRate             string                       `json:"interest_rate"`
	InterestPaymentSchedule  string                       `json:"interest_payment_schedule"`
	PrincipalPaymentSchedule string                       `json:"principal_payment_schedule"`
	UseOfFunds               []ProjectUseOfFunds          `json:"use_of_funds"`
	CollateralGuarantee      []ProjectCollateralGuarantee `json:"collateral_guarantee"`
	DescJob                  string                       `json:"desc_job"`
	IsApbn                   bool                         `json:"is_apbn"`
	IsApproved               bool                         `json:"is_approved"`
	CanRefund                bool                         `json:"can_refund"`
	MulaiProject             string                       `json:"mulai_project"`
	SelesaiProject           string                       `json:"selesai_project"`
	AlamatPenyediaProject    string                       `json:"alamat_penyedia_project"`
	AlamatPenyediaProvinsi   string                       `json:"alamat_penyedia_provinsi"`
	AlamatPenyediaKota       string                       `json:"alamat_penyedia_kota"`
	AlamatPenyediaDaerah     string                       `json:"alamat_penyedia_daerah"`
	AlamatPenyediaWilayah    string                       `json:"alamat_penyedia_wilayah"`
	AlamatPenyediaKodePos    int64                        `json:"alamat_penyedia_kode_pos"`
	RemainingDays            int                          `json:"remaining_days"`
	ProjectIsExpire          bool                         `json:"project_is_expire"`
	Status                   string                       `json:"status"`
	Company                  Company                      `json:"company"`
	CreatedAt                time.Time                    `json:"created_at"`
	UpdatedAt                time.Time                    `json:"updated_at"`
}

type ProjectUseOfFunds struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ProjectCollateralGuarantee struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Company struct {
	Name       string `json:"name"`
	JenisUsaha string `json:"jenis_usaha"`
}

type ProjectStore struct {
	Id                          string                    `json:"id"`
	CompanyId                   string                    `json:"company_id"`
	Title                       string                    `json:"title"`
	Deskripsi                   string                    `json:"deskripsi"`
	TingkatBunga                string                    `json:"tingkat_bunga"`
	JenisProject                string                    `json:"jenis_project"`
	AlamatPenyediaProject       string                    `json:"alamat_penyedia_project"`
	AlamatPenyediaProvinsi      string                    `json:"alamat_penyedia_provinsi"`
	AlamatPenyediaKota          string                    `json:"alamat_penyedia_kota"`
	AlamatPenyediaDaerah        string                    `json:"alamat_penyedia_daerah"`
	AlamatPenyediaWilayah       string                    `json:"alamat_penyedia_wilayah"`
	AlamatPenyediaKodePos       string                    `json:"alamat_penyedia_kode_pos"`
	MulaiProject                string                    `json:"mulai_project"`
	SelesaiProject              string                    `json:"selesai_project"`
	JangkaWaktu                 string                    `json:"jangka_waktu"`
	JumlahMinimal               string                    `json:"jumlah_minimal"`
	JadwalPembayaranBunga       string                    `json:"jadwal_pembayaran_bunga"`
	JadwalPembayaranPokok       string                    `json:"jadwal_pembayaran_pokok"`
	PersentaseKeuntungan        string                    `json:"persentase_keuntungan"`
	Modal                       string                    `json:"modal"`
	Spk                         string                    `json:"spk"`
	Loa                         string                    `json:"loa"`
	DanaYangDibutuhkan          int64                     `json:"dana_yang_dibutuhkan"`
	JumlahLot                   string                    `json:"jumlah_lot"`
	CompanyProfile              string                    `json:"company_profile"`
	NoContractValue             string                    `json:"no_contract_value"`
	NoContractPath              string                    `json:"no_contract_path"`
	PenggunaanDana              []ProjectPenggunaanDana   `json:"penggunaan_dana"`
	JaminanKolateral            []ProjectJaminanKolateral `json:"jaminan_kolateral"`
	Media                       []ProjectMedia            `json:"media"`
	BatasAkhirPengerjaan        string                    `json:"batas_akhir_pengerjaan"`
	Website                     string                    `json:"website"`
	TenorPinjaman               string                    `json:"tenor_pinjaman"`
	InstansiPemberiProject      string                    `json:"instansi_pemberi_project"`
	JenisInstansiPemberiProject string                    `json:"jenis_instansi_pemberi_project"`
	IsApbn                      bool                      `json:"is_apbn"`
	DocRekeningKoran            string                    `json:"doc_rekening_koran"`
	DocLaporanKeuangan          string                    `json:"doc_laporan_keuangan"`
	DocContract                 string                    `json:"doc_contract"`
	DocProspect                 string                    `json:"doc_prospect"`
	Location                    ProjectLocation           `json:"location"`
}

type ProjectPenggunaanDana struct {
	Name string `json:"name"`
}

type ProjectJaminanKolateral struct {
	Name string `json:"name"`
}

type ProjectMedia struct {
	Id   int    `json:"id"`
	Path string `json:"path"`
}

type ProjectStoreMedia struct {
	Id        string `json:"id"`
	Path      string `json:"path"`
	ProjectId string `json:"project_id"`
}

type ProjectLocation struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
	Lat  string `json:"lat"`
	Lng  string `json:"lng"`
}

type ProjectStoreLocation struct {
	Id        int    `json:"id"`
	ProjectId string `json:"project_id"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Lat       string `json:"lat"`
	Lng       string `json:"lng"`
}

type ProjectDoc struct {
	Id   string `json:"id"`
	Path string `json:"path"`
}

type ProjectCompany struct {
	Id              string `json:"id"`
	CompanyName     string `json:"company_name"`
	JenisUsaha      string `json:"jenis_usaha"`
	Address         string `json:"address"`
	PostalCode      string `json:"postal_code"`
	ProvinceName    string `json:"province_name"`
	CityName        string `json:"city_name"`
	DistrictName    string `json:"district_name"`
	SubdistrictName string `json:"subdistrict_name"`
}

type ProjectUpdate struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Goal      string `json:"goal"`
	Capital   string `json:"capital"`
	MinInvest string `json:"min_invest"`
	UnitPrice string `json:"unit_price"`
	UnitTotal string `json:"unit_total"`
}

type ProjectPayment struct {
	IsApprove bool   `json:"is_approve"`
	Path      string `json:"path"`
}

type ProjectDocumentVerify struct {
	Id                      string `json:"id"`
	Skd                     string `json:"skd"`
	Cv                      string `json:"cv"`
	Rab                     string `json:"rab"`
	CashflowProject         string `json:"cashflow_project"`
	OtherLicenseDocument    string `json:"other_license_document"`
	VideoProfileCompany     string `json:"video_profile_company"`
	ProjectSummary          string `json:"project_summary"`
	RevenueProjection       string `json:"revenue_projection"`
	WorkOfTimeline          string `json:"work_of_timeline"`
	AnnualTaxReport         string `json:"annual_tax_report"`
	ListOfEmployment        string `json:"list_of_employment"`
	ListOfSupplierData      string `json:"list_of_supplier_data"`
	LatestReceivableAccount string `json:"latest_receivable_account"`

	Media []MediaDocumentVerifyProject `json:"media" gorm:"-"`
}

type MediaDocumentVerifyProject struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type ProjectCostOfFund struct {
	ProjectId string `json:"project_id"`
}

type ProjectInquiry struct {
	ProjectId string `json:"project_id"`
}

type ProjectCostOfFundPayment struct {
	CostOfFundId string `json:"cost_of_fund_id"`
	PaymentNo    string `json:"payment_no"`
	Amount       string `json:"amount"`
	PaymentDate  string `json:"payment_date"`
}

type ProjectDelete struct {
	Id string `json:"id"`
}
