// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	CountryCode string
)

const (
	CountryCodeAD CountryCode = "AD"
	CountryCodeAE CountryCode = "AE"
	CountryCodeAF CountryCode = "AF"
	CountryCodeAG CountryCode = "AG"
	CountryCodeAI CountryCode = "AI"
	CountryCodeAL CountryCode = "AL"
	CountryCodeAM CountryCode = "AM"
	CountryCodeAO CountryCode = "AO"
	CountryCodeAQ CountryCode = "AQ"
	CountryCodeAR CountryCode = "AR"
	CountryCodeAS CountryCode = "AS"
	CountryCodeAT CountryCode = "AT"
	CountryCodeAU CountryCode = "AU"
	CountryCodeAW CountryCode = "AW"
	CountryCodeAX CountryCode = "AX"
	CountryCodeAZ CountryCode = "AZ"
	CountryCodeBA CountryCode = "BA"
	CountryCodeBB CountryCode = "BB"
	CountryCodeBD CountryCode = "BD"
	CountryCodeBE CountryCode = "BE"
	CountryCodeBF CountryCode = "BF"
	CountryCodeBG CountryCode = "BG"
	CountryCodeBH CountryCode = "BH"
	CountryCodeBI CountryCode = "BI"
	CountryCodeBJ CountryCode = "BJ"
	CountryCodeBL CountryCode = "BL"
	CountryCodeBM CountryCode = "BM"
	CountryCodeBN CountryCode = "BN"
	CountryCodeBO CountryCode = "BO"
	CountryCodeBQ CountryCode = "BQ"
	CountryCodeBR CountryCode = "BR"
	CountryCodeBS CountryCode = "BS"
	CountryCodeBT CountryCode = "BT"
	CountryCodeBV CountryCode = "BV"
	CountryCodeBW CountryCode = "BW"
	CountryCodeBY CountryCode = "BY"
	CountryCodeBZ CountryCode = "BZ"
	CountryCodeCA CountryCode = "CA"
	CountryCodeCC CountryCode = "CC"
	CountryCodeCD CountryCode = "CD"
	CountryCodeCF CountryCode = "CF"
	CountryCodeCG CountryCode = "CG"
	CountryCodeCH CountryCode = "CH"
	CountryCodeCI CountryCode = "CI"
	CountryCodeCK CountryCode = "CK"
	CountryCodeCL CountryCode = "CL"
	CountryCodeCM CountryCode = "CM"
	CountryCodeCN CountryCode = "CN"
	CountryCodeCO CountryCode = "CO"
	CountryCodeCR CountryCode = "CR"
	CountryCodeCU CountryCode = "CU"
	CountryCodeCV CountryCode = "CV"
	CountryCodeCW CountryCode = "CW"
	CountryCodeCX CountryCode = "CX"
	CountryCodeCY CountryCode = "CY"
	CountryCodeCZ CountryCode = "CZ"
	CountryCodeDE CountryCode = "DE"
	CountryCodeDJ CountryCode = "DJ"
	CountryCodeDK CountryCode = "DK"
	CountryCodeDM CountryCode = "DM"
	CountryCodeDO CountryCode = "DO"
	CountryCodeDZ CountryCode = "DZ"
	CountryCodeEC CountryCode = "EC"
	CountryCodeEE CountryCode = "EE"
	CountryCodeEG CountryCode = "EG"
	CountryCodeEH CountryCode = "EH"
	CountryCodeER CountryCode = "ER"
	CountryCodeES CountryCode = "ES"
	CountryCodeET CountryCode = "ET"
	CountryCodeEU CountryCode = "EU"
	CountryCodeFI CountryCode = "FI"
	CountryCodeFJ CountryCode = "FJ"
	CountryCodeFK CountryCode = "FK"
	CountryCodeFM CountryCode = "FM"
	CountryCodeFO CountryCode = "FO"
	CountryCodeFR CountryCode = "FR"
	CountryCodeGA CountryCode = "GA"
	CountryCodeGB CountryCode = "GB"
	CountryCodeGD CountryCode = "GD"
	CountryCodeGE CountryCode = "GE"
	CountryCodeGF CountryCode = "GF"
	CountryCodeGG CountryCode = "GG"
	CountryCodeGH CountryCode = "GH"
	CountryCodeGI CountryCode = "GI"
	CountryCodeGL CountryCode = "GL"
	CountryCodeGM CountryCode = "GM"
	CountryCodeGN CountryCode = "GN"
	CountryCodeGP CountryCode = "GP"
	CountryCodeGQ CountryCode = "GQ"
	CountryCodeGR CountryCode = "GR"
	CountryCodeGT CountryCode = "GT"
	CountryCodeGU CountryCode = "GU"
	CountryCodeGW CountryCode = "GW"
	CountryCodeGY CountryCode = "GY"
	CountryCodeHK CountryCode = "HK"
	CountryCodeHM CountryCode = "HM"
	CountryCodeHN CountryCode = "HN"
	CountryCodeHR CountryCode = "HR"
	CountryCodeHT CountryCode = "HT"
	CountryCodeHU CountryCode = "HU"
	CountryCodeID CountryCode = "ID"
	CountryCodeIE CountryCode = "IE"
	CountryCodeIL CountryCode = "IL"
	CountryCodeIM CountryCode = "IM"
	CountryCodeIN CountryCode = "IN"
	CountryCodeIO CountryCode = "IO"
	CountryCodeIQ CountryCode = "IQ"
	CountryCodeIR CountryCode = "IR"
	CountryCodeIS CountryCode = "IS"
	CountryCodeIT CountryCode = "IT"
	CountryCodeJE CountryCode = "JE"
	CountryCodeJM CountryCode = "JM"
	CountryCodeJO CountryCode = "JO"
	CountryCodeJP CountryCode = "JP"
	CountryCodeKE CountryCode = "KE"
	CountryCodeKG CountryCode = "KG"
	CountryCodeKH CountryCode = "KH"
	CountryCodeKI CountryCode = "KI"
	CountryCodeKM CountryCode = "KM"
	CountryCodeKN CountryCode = "KN"
	CountryCodeKP CountryCode = "KP"
	CountryCodeKR CountryCode = "KR"
	CountryCodeKW CountryCode = "KW"
	CountryCodeKY CountryCode = "KY"
	CountryCodeKZ CountryCode = "KZ"
	CountryCodeLA CountryCode = "LA"
	CountryCodeLB CountryCode = "LB"
	CountryCodeLC CountryCode = "LC"
	CountryCodeLI CountryCode = "LI"
	CountryCodeLK CountryCode = "LK"
	CountryCodeLR CountryCode = "LR"
	CountryCodeLS CountryCode = "LS"
	CountryCodeLT CountryCode = "LT"
	CountryCodeLU CountryCode = "LU"
	CountryCodeLV CountryCode = "LV"
	CountryCodeLY CountryCode = "LY"
	CountryCodeMA CountryCode = "MA"
	CountryCodeMC CountryCode = "MC"
	CountryCodeMD CountryCode = "MD"
	CountryCodeME CountryCode = "ME"
	CountryCodeMF CountryCode = "MF"
	CountryCodeMG CountryCode = "MG"
	CountryCodeMH CountryCode = "MH"
	CountryCodeMK CountryCode = "MK"
	CountryCodeML CountryCode = "ML"
	CountryCodeMM CountryCode = "MM"
	CountryCodeMN CountryCode = "MN"
	CountryCodeMO CountryCode = "MO"
	CountryCodeMP CountryCode = "MP"
	CountryCodeMQ CountryCode = "MQ"
	CountryCodeMR CountryCode = "MR"
	CountryCodeMS CountryCode = "MS"
	CountryCodeMT CountryCode = "MT"
	CountryCodeMU CountryCode = "MU"
	CountryCodeMV CountryCode = "MV"
	CountryCodeMW CountryCode = "MW"
	CountryCodeMX CountryCode = "MX"
	CountryCodeMY CountryCode = "MY"
	CountryCodeMZ CountryCode = "MZ"
	CountryCodeNA CountryCode = "NA"
	CountryCodeNC CountryCode = "NC"
	CountryCodeNE CountryCode = "NE"
	CountryCodeNF CountryCode = "NF"
	CountryCodeNG CountryCode = "NG"
	CountryCodeNI CountryCode = "NI"
	CountryCodeNL CountryCode = "NL"
	CountryCodeNO CountryCode = "NO"
	CountryCodeNP CountryCode = "NP"
	CountryCodeNR CountryCode = "NR"
	CountryCodeNU CountryCode = "NU"
	CountryCodeNZ CountryCode = "NZ"
	CountryCodeOM CountryCode = "OM"
	CountryCodePA CountryCode = "PA"
	CountryCodePE CountryCode = "PE"
	CountryCodePF CountryCode = "PF"
	CountryCodePG CountryCode = "PG"
	CountryCodePH CountryCode = "PH"
	CountryCodePK CountryCode = "PK"
	CountryCodePL CountryCode = "PL"
	CountryCodePM CountryCode = "PM"
	CountryCodePN CountryCode = "PN"
	CountryCodePR CountryCode = "PR"
	CountryCodePS CountryCode = "PS"
	CountryCodePT CountryCode = "PT"
	CountryCodePW CountryCode = "PW"
	CountryCodePY CountryCode = "PY"
	CountryCodeQA CountryCode = "QA"
	CountryCodeRE CountryCode = "RE"
	CountryCodeRO CountryCode = "RO"
	CountryCodeRS CountryCode = "RS"
	CountryCodeRU CountryCode = "RU"
	CountryCodeRW CountryCode = "RW"
	CountryCodeSA CountryCode = "SA"
	CountryCodeSB CountryCode = "SB"
	CountryCodeSC CountryCode = "SC"
	CountryCodeSD CountryCode = "SD"
	CountryCodeSE CountryCode = "SE"
	CountryCodeSG CountryCode = "SG"
	CountryCodeSH CountryCode = "SH"
	CountryCodeSI CountryCode = "SI"
	CountryCodeSJ CountryCode = "SJ"
	CountryCodeSK CountryCode = "SK"
	CountryCodeSL CountryCode = "SL"
	CountryCodeSM CountryCode = "SM"
	CountryCodeSN CountryCode = "SN"
	CountryCodeSO CountryCode = "SO"
	CountryCodeSR CountryCode = "SR"
	CountryCodeSS CountryCode = "SS"
	CountryCodeST CountryCode = "ST"
	CountryCodeSV CountryCode = "SV"
	CountryCodeSX CountryCode = "SX"
	CountryCodeSY CountryCode = "SY"
	CountryCodeSZ CountryCode = "SZ"
	CountryCodeTC CountryCode = "TC"
	CountryCodeTD CountryCode = "TD"
	CountryCodeTF CountryCode = "TF"
	CountryCodeTG CountryCode = "TG"
	CountryCodeTH CountryCode = "TH"
	CountryCodeTJ CountryCode = "TJ"
	CountryCodeTK CountryCode = "TK"
	CountryCodeTL CountryCode = "TL"
	CountryCodeTM CountryCode = "TM"
	CountryCodeTN CountryCode = "TN"
	CountryCodeTO CountryCode = "TO"
	CountryCodeTR CountryCode = "TR"
	CountryCodeTT CountryCode = "TT"
	CountryCodeTV CountryCode = "TV"
	CountryCodeTW CountryCode = "TW"
	CountryCodeTZ CountryCode = "TZ"
	CountryCodeUA CountryCode = "UA"
	CountryCodeUG CountryCode = "UG"
	CountryCodeUM CountryCode = "UM"
	CountryCodeUS CountryCode = "US"
	CountryCodeUY CountryCode = "UY"
	CountryCodeUZ CountryCode = "UZ"
	CountryCodeVA CountryCode = "VA"
	CountryCodeVC CountryCode = "VC"
	CountryCodeVE CountryCode = "VE"
	CountryCodeVG CountryCode = "VG"
	CountryCodeVI CountryCode = "VI"
	CountryCodeVN CountryCode = "VN"
	CountryCodeVU CountryCode = "VU"
	CountryCodeWF CountryCode = "WF"
	CountryCodeWS CountryCode = "WS"
	CountryCodeYE CountryCode = "YE"
	CountryCodeYT CountryCode = "YT"
	CountryCodeZA CountryCode = "ZA"
	CountryCodeZM CountryCode = "ZM"
	CountryCodeZW CountryCode = "ZW"
)

func (ct CountryCode) String() string {
	return string(ct)
}

func (ct *CountryCode) Scan(value any) error {
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for CountryCode: %T", value)
	}

	switch s {
	case CountryCodeAD.String():
		*ct = CountryCodeAD
	case CountryCodeAE.String():
		*ct = CountryCodeAE
	case CountryCodeAF.String():
		*ct = CountryCodeAF
	case CountryCodeAG.String():
		*ct = CountryCodeAG
	case CountryCodeAI.String():
		*ct = CountryCodeAI
	case CountryCodeAL.String():
		*ct = CountryCodeAL
	case CountryCodeAM.String():
		*ct = CountryCodeAM
	case CountryCodeAO.String():
		*ct = CountryCodeAO
	case CountryCodeAQ.String():
		*ct = CountryCodeAQ
	case CountryCodeAR.String():
		*ct = CountryCodeAR
	case CountryCodeAS.String():
		*ct = CountryCodeAS
	case CountryCodeAT.String():
		*ct = CountryCodeAT
	case CountryCodeAU.String():
		*ct = CountryCodeAU
	case CountryCodeAW.String():
		*ct = CountryCodeAW
	case CountryCodeAX.String():
		*ct = CountryCodeAX
	case CountryCodeAZ.String():
		*ct = CountryCodeAZ
	case CountryCodeBA.String():
		*ct = CountryCodeBA
	case CountryCodeBB.String():
		*ct = CountryCodeBB
	case CountryCodeBD.String():
		*ct = CountryCodeBD
	case CountryCodeBE.String():
		*ct = CountryCodeBE
	case CountryCodeBF.String():
		*ct = CountryCodeBF
	case CountryCodeBG.String():
		*ct = CountryCodeBG
	case CountryCodeBH.String():
		*ct = CountryCodeBH
	case CountryCodeBI.String():
		*ct = CountryCodeBI
	case CountryCodeBJ.String():
		*ct = CountryCodeBJ
	case CountryCodeBL.String():
		*ct = CountryCodeBL
	case CountryCodeBM.String():
		*ct = CountryCodeBM
	case CountryCodeBN.String():
		*ct = CountryCodeBN
	case CountryCodeBO.String():
		*ct = CountryCodeBO
	case CountryCodeBQ.String():
		*ct = CountryCodeBQ
	case CountryCodeBR.String():
		*ct = CountryCodeBR
	case CountryCodeBS.String():
		*ct = CountryCodeBS
	case CountryCodeBT.String():
		*ct = CountryCodeBT
	case CountryCodeBV.String():
		*ct = CountryCodeBV
	case CountryCodeBW.String():
		*ct = CountryCodeBW
	case CountryCodeBY.String():
		*ct = CountryCodeBY
	case CountryCodeBZ.String():
		*ct = CountryCodeBZ
	case CountryCodeCA.String():
		*ct = CountryCodeCA
	case CountryCodeCC.String():
		*ct = CountryCodeCC
	case CountryCodeCD.String():
		*ct = CountryCodeCD
	case CountryCodeCF.String():
		*ct = CountryCodeCF
	case CountryCodeCG.String():
		*ct = CountryCodeCG
	case CountryCodeCH.String():
		*ct = CountryCodeCH
	case CountryCodeCI.String():
		*ct = CountryCodeCI
	case CountryCodeCK.String():
		*ct = CountryCodeCK
	case CountryCodeCL.String():
		*ct = CountryCodeCL
	case CountryCodeCM.String():
		*ct = CountryCodeCM
	case CountryCodeCN.String():
		*ct = CountryCodeCN
	case CountryCodeCO.String():
		*ct = CountryCodeCO
	case CountryCodeCR.String():
		*ct = CountryCodeCR
	case CountryCodeCU.String():
		*ct = CountryCodeCU
	case CountryCodeCV.String():
		*ct = CountryCodeCV
	case CountryCodeCW.String():
		*ct = CountryCodeCW
	case CountryCodeCX.String():
		*ct = CountryCodeCX
	case CountryCodeCY.String():
		*ct = CountryCodeCY
	case CountryCodeCZ.String():
		*ct = CountryCodeCZ
	case CountryCodeDE.String():
		*ct = CountryCodeDE
	case CountryCodeDJ.String():
		*ct = CountryCodeDJ
	case CountryCodeDK.String():
		*ct = CountryCodeDK
	case CountryCodeDM.String():
		*ct = CountryCodeDM
	case CountryCodeDO.String():
		*ct = CountryCodeDO
	case CountryCodeDZ.String():
		*ct = CountryCodeDZ
	case CountryCodeEC.String():
		*ct = CountryCodeEC
	case CountryCodeEE.String():
		*ct = CountryCodeEE
	case CountryCodeEG.String():
		*ct = CountryCodeEG
	case CountryCodeEH.String():
		*ct = CountryCodeEH
	case CountryCodeER.String():
		*ct = CountryCodeER
	case CountryCodeES.String():
		*ct = CountryCodeES
	case CountryCodeET.String():
		*ct = CountryCodeET
	case CountryCodeEU.String():
		*ct = CountryCodeEU
	case CountryCodeFI.String():
		*ct = CountryCodeFI
	case CountryCodeFJ.String():
		*ct = CountryCodeFJ
	case CountryCodeFK.String():
		*ct = CountryCodeFK
	case CountryCodeFM.String():
		*ct = CountryCodeFM
	case CountryCodeFO.String():
		*ct = CountryCodeFO
	case CountryCodeFR.String():
		*ct = CountryCodeFR
	case CountryCodeGA.String():
		*ct = CountryCodeGA
	case CountryCodeGB.String():
		*ct = CountryCodeGB
	case CountryCodeGD.String():
		*ct = CountryCodeGD
	case CountryCodeGE.String():
		*ct = CountryCodeGE
	case CountryCodeGF.String():
		*ct = CountryCodeGF
	case CountryCodeGG.String():
		*ct = CountryCodeGG
	case CountryCodeGH.String():
		*ct = CountryCodeGH
	case CountryCodeGI.String():
		*ct = CountryCodeGI
	case CountryCodeGL.String():
		*ct = CountryCodeGL
	case CountryCodeGM.String():
		*ct = CountryCodeGM
	case CountryCodeGN.String():
		*ct = CountryCodeGN
	case CountryCodeGP.String():
		*ct = CountryCodeGP
	case CountryCodeGQ.String():
		*ct = CountryCodeGQ
	case CountryCodeGR.String():
		*ct = CountryCodeGR
	case CountryCodeGT.String():
		*ct = CountryCodeGT
	case CountryCodeGU.String():
		*ct = CountryCodeGU
	case CountryCodeGW.String():
		*ct = CountryCodeGW
	case CountryCodeGY.String():
		*ct = CountryCodeGY
	case CountryCodeHK.String():
		*ct = CountryCodeHK
	case CountryCodeHM.String():
		*ct = CountryCodeHM
	case CountryCodeHN.String():
		*ct = CountryCodeHN
	case CountryCodeHR.String():
		*ct = CountryCodeHR
	case CountryCodeHT.String():
		*ct = CountryCodeHT
	case CountryCodeHU.String():
		*ct = CountryCodeHU
	case CountryCodeID.String():
		*ct = CountryCodeID
	case CountryCodeIE.String():
		*ct = CountryCodeIE
	case CountryCodeIL.String():
		*ct = CountryCodeIL
	case CountryCodeIM.String():
		*ct = CountryCodeIM
	case CountryCodeIN.String():
		*ct = CountryCodeIN
	case CountryCodeIO.String():
		*ct = CountryCodeIO
	case CountryCodeIQ.String():
		*ct = CountryCodeIQ
	case CountryCodeIR.String():
		*ct = CountryCodeIR
	case CountryCodeIS.String():
		*ct = CountryCodeIS
	case CountryCodeIT.String():
		*ct = CountryCodeIT
	case CountryCodeJE.String():
		*ct = CountryCodeJE
	case CountryCodeJM.String():
		*ct = CountryCodeJM
	case CountryCodeJO.String():
		*ct = CountryCodeJO
	case CountryCodeJP.String():
		*ct = CountryCodeJP
	case CountryCodeKE.String():
		*ct = CountryCodeKE
	case CountryCodeKG.String():
		*ct = CountryCodeKG
	case CountryCodeKH.String():
		*ct = CountryCodeKH
	case CountryCodeKI.String():
		*ct = CountryCodeKI
	case CountryCodeKM.String():
		*ct = CountryCodeKM
	case CountryCodeKN.String():
		*ct = CountryCodeKN
	case CountryCodeKP.String():
		*ct = CountryCodeKP
	case CountryCodeKR.String():
		*ct = CountryCodeKR
	case CountryCodeKW.String():
		*ct = CountryCodeKW
	case CountryCodeKY.String():
		*ct = CountryCodeKY
	case CountryCodeKZ.String():
		*ct = CountryCodeKZ
	case CountryCodeLA.String():
		*ct = CountryCodeLA
	case CountryCodeLB.String():
		*ct = CountryCodeLB
	case CountryCodeLC.String():
		*ct = CountryCodeLC
	case CountryCodeLI.String():
		*ct = CountryCodeLI
	case CountryCodeLK.String():
		*ct = CountryCodeLK
	case CountryCodeLR.String():
		*ct = CountryCodeLR
	case CountryCodeLS.String():
		*ct = CountryCodeLS
	case CountryCodeLT.String():
		*ct = CountryCodeLT
	case CountryCodeLU.String():
		*ct = CountryCodeLU
	case CountryCodeLV.String():
		*ct = CountryCodeLV
	case CountryCodeLY.String():
		*ct = CountryCodeLY
	case CountryCodeMA.String():
		*ct = CountryCodeMA
	case CountryCodeMC.String():
		*ct = CountryCodeMC
	case CountryCodeMD.String():
		*ct = CountryCodeMD
	case CountryCodeME.String():
		*ct = CountryCodeME
	case CountryCodeMF.String():
		*ct = CountryCodeMF
	case CountryCodeMG.String():
		*ct = CountryCodeMG
	case CountryCodeMH.String():
		*ct = CountryCodeMH
	case CountryCodeMK.String():
		*ct = CountryCodeMK
	case CountryCodeML.String():
		*ct = CountryCodeML
	case CountryCodeMM.String():
		*ct = CountryCodeMM
	case CountryCodeMN.String():
		*ct = CountryCodeMN
	case CountryCodeMO.String():
		*ct = CountryCodeMO
	case CountryCodeMP.String():
		*ct = CountryCodeMP
	case CountryCodeMQ.String():
		*ct = CountryCodeMQ
	case CountryCodeMR.String():
		*ct = CountryCodeMR
	case CountryCodeMS.String():
		*ct = CountryCodeMS
	case CountryCodeMT.String():
		*ct = CountryCodeMT
	case CountryCodeMU.String():
		*ct = CountryCodeMU
	case CountryCodeMV.String():
		*ct = CountryCodeMV
	case CountryCodeMW.String():
		*ct = CountryCodeMW
	case CountryCodeMX.String():
		*ct = CountryCodeMX
	case CountryCodeMY.String():
		*ct = CountryCodeMY
	case CountryCodeMZ.String():
		*ct = CountryCodeMZ
	case CountryCodeNA.String():
		*ct = CountryCodeNA
	case CountryCodeNC.String():
		*ct = CountryCodeNC
	case CountryCodeNE.String():
		*ct = CountryCodeNE
	case CountryCodeNF.String():
		*ct = CountryCodeNF
	case CountryCodeNG.String():
		*ct = CountryCodeNG
	case CountryCodeNI.String():
		*ct = CountryCodeNI
	case CountryCodeNL.String():
		*ct = CountryCodeNL
	case CountryCodeNO.String():
		*ct = CountryCodeNO
	case CountryCodeNP.String():
		*ct = CountryCodeNP
	case CountryCodeNR.String():
		*ct = CountryCodeNR
	case CountryCodeNU.String():
		*ct = CountryCodeNU
	case CountryCodeNZ.String():
		*ct = CountryCodeNZ
	case CountryCodeOM.String():
		*ct = CountryCodeOM
	case CountryCodePA.String():
		*ct = CountryCodePA
	case CountryCodePE.String():
		*ct = CountryCodePE
	case CountryCodePF.String():
		*ct = CountryCodePF
	case CountryCodePG.String():
		*ct = CountryCodePG
	case CountryCodePH.String():
		*ct = CountryCodePH
	case CountryCodePK.String():
		*ct = CountryCodePK
	case CountryCodePL.String():
		*ct = CountryCodePL
	case CountryCodePM.String():
		*ct = CountryCodePM
	case CountryCodePN.String():
		*ct = CountryCodePN
	case CountryCodePR.String():
		*ct = CountryCodePR
	case CountryCodePS.String():
		*ct = CountryCodePS
	case CountryCodePT.String():
		*ct = CountryCodePT
	case CountryCodePW.String():
		*ct = CountryCodePW
	case CountryCodePY.String():
		*ct = CountryCodePY
	case CountryCodeQA.String():
		*ct = CountryCodeQA
	case CountryCodeRE.String():
		*ct = CountryCodeRE
	case CountryCodeRO.String():
		*ct = CountryCodeRO
	case CountryCodeRS.String():
		*ct = CountryCodeRS
	case CountryCodeRU.String():
		*ct = CountryCodeRU
	case CountryCodeRW.String():
		*ct = CountryCodeRW
	case CountryCodeSA.String():
		*ct = CountryCodeSA
	case CountryCodeSB.String():
		*ct = CountryCodeSB
	case CountryCodeSC.String():
		*ct = CountryCodeSC
	case CountryCodeSD.String():
		*ct = CountryCodeSD
	case CountryCodeSE.String():
		*ct = CountryCodeSE
	case CountryCodeSG.String():
		*ct = CountryCodeSG
	case CountryCodeSH.String():
		*ct = CountryCodeSH
	case CountryCodeSI.String():
		*ct = CountryCodeSI
	case CountryCodeSJ.String():
		*ct = CountryCodeSJ
	case CountryCodeSK.String():
		*ct = CountryCodeSK
	case CountryCodeSL.String():
		*ct = CountryCodeSL
	case CountryCodeSM.String():
		*ct = CountryCodeSM
	case CountryCodeSN.String():
		*ct = CountryCodeSN
	case CountryCodeSO.String():
		*ct = CountryCodeSO
	case CountryCodeSR.String():
		*ct = CountryCodeSR
	case CountryCodeSS.String():
		*ct = CountryCodeSS
	case CountryCodeST.String():
		*ct = CountryCodeST
	case CountryCodeSV.String():
		*ct = CountryCodeSV
	case CountryCodeSY.String():
		*ct = CountryCodeSY
	case CountryCodeSZ.String():
		*ct = CountryCodeSZ
	case CountryCodeTC.String():
		*ct = CountryCodeTC
	case CountryCodeTD.String():
		*ct = CountryCodeTD
	case CountryCodeTF.String():
		*ct = CountryCodeTF
	case CountryCodeTG.String():
		*ct = CountryCodeTG
	case CountryCodeTH.String():
		*ct = CountryCodeTH
	case CountryCodeTJ.String():
		*ct = CountryCodeTJ
	case CountryCodeTK.String():
		*ct = CountryCodeTK
	case CountryCodeTL.String():
		*ct = CountryCodeTL
	case CountryCodeTM.String():
		*ct = CountryCodeTM
	case CountryCodeTN.String():
		*ct = CountryCodeTN
	case CountryCodeTO.String():
		*ct = CountryCodeTO
	case CountryCodeTR.String():
		*ct = CountryCodeTR
	case CountryCodeTT.String():
		*ct = CountryCodeTT
	case CountryCodeTV.String():
		*ct = CountryCodeTV
	case CountryCodeTW.String():
		*ct = CountryCodeTW
	case CountryCodeTZ.String():
		*ct = CountryCodeTZ
	case CountryCodeUA.String():
		*ct = CountryCodeUA
	case CountryCodeUG.String():
		*ct = CountryCodeUG
	case CountryCodeUM.String():
		*ct = CountryCodeUM
	case CountryCodeUS.String():
		*ct = CountryCodeUS
	case CountryCodeUY.String():
		*ct = CountryCodeUY
	case CountryCodeUZ.String():
		*ct = CountryCodeUZ
	case CountryCodeVA.String():
		*ct = CountryCodeVA
	case CountryCodeVC.String():
		*ct = CountryCodeVC
	case CountryCodeVE.String():
		*ct = CountryCodeVE
	case CountryCodeVG.String():
		*ct = CountryCodeVG
	case CountryCodeVI.String():
		*ct = CountryCodeVI
	case CountryCodeVN.String():
		*ct = CountryCodeVN
	case CountryCodeVU.String():
		*ct = CountryCodeVU
	case CountryCodeWF.String():
		*ct = CountryCodeWF
	case CountryCodeWS.String():
		*ct = CountryCodeWS
	case CountryCodeYE.String():
		*ct = CountryCodeYE
	case CountryCodeYT.String():
		*ct = CountryCodeYT
	case CountryCodeZA.String():
		*ct = CountryCodeZA
	case CountryCodeZM.String():
		*ct = CountryCodeZM
	case CountryCodeZW.String():
		*ct = CountryCodeZW
	default:
		return fmt.Errorf("invalid CountryCode value: %q", s)
	}
	return nil
}

func (st CountryCode) Value() (driver.Value, error) {
	return st.String(), nil
}

type CountryCodes []CountryCode

func (s *CountryCodes) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return s.scanFromString(v)
	case []byte:
		return s.scanFromString(string(v))
	default:
		return fmt.Errorf("unsupported type for CountryCodes: %T", value)
	}
}

func (s *CountryCodes) scanFromString(str string) error {
	str = strings.TrimSpace(str)
	if str == "{}" || str == "" {
		*s = []CountryCode{}
		return nil
	}

	if strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}") {
		str = str[1 : len(str)-1]
	}

	parts := strings.Split(str, ",")
	result := make([]CountryCode, len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)

		if strings.HasPrefix(part, `"`) && strings.HasSuffix(part, `"`) {
			part = part[1 : len(part)-1]
		}

		var ct CountryCode
		if err := ct.Scan(part); err != nil {
			return fmt.Errorf("invalid country code in array: %s", part)
		}
		result[i] = ct
	}

	*s = result
	return nil
}

func (s CountryCodes) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "{}", nil
	}

	values := make([]string, len(s))
	for i, ct := range s {
		values[i] = ct.String()
	}

	return "{" + strings.Join(values, ",") + "}", nil
}
