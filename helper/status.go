package helper

func ResolveApprovalActivityStatus(role string) string {
	switch role {
	case "admin":
		return "APPROVED"
	case "sekretaris departemen":
		return "APPROVED"
	case "dosen pembimbing":
		return "PENDING"
	case "dosen pemonev":
		return "PENDING"
	case "LO - MBKM":
		return "APPROVED"
	case "mahasiswa":
		return "PENDING"
	default:
		return "PENDING"
	}
}
