package members

import (
	"encoding/csv"
	"fmt"
	"go-scrape-redmine/config"
	"go-scrape-redmine/models"
	"go-scrape-redmine/seed"
	"os"
	"io"
	"bufio"
)

func NewMember() seed.Member {
	return &Member{}
}

type Member struct{}

type memberData struct {
	MemberID    string
	MemberName  string
	MemberEmail string
}

func readCSV(rs io.ReadSeeker) ([][]string, error) {
    // Skip first row (line)
    row1, err := bufio.NewReader(rs).ReadSlice('\n')
    if err != nil {
        return nil, err
    }
    _, err = rs.Seek(int64(len(row1)), io.SeekStart)
    if err != nil {
        return nil, err
    }

    // Read remaining rows
    r := csv.NewReader(rs)
    rows, err := r.ReadAll()
    if err != nil {
        return nil, err
    }
    return rows, nil
}
func (a *Member) SeedMember() {
	db := config.DBConnect()
	csvFile, err := os.Open(os.Getenv("SEED_FILE_PATH"))
	if err != nil {
		fmt.Println("members.csv not found")
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened CSV file")
	defer csvFile.Close()

	csvLines, err := readCSV(csvFile)
	if err != nil {
		fmt.Println(err)
	}
	for _, line := range csvLines {
		memberdata := memberData{
			MemberID:    line[0],
			MemberName:  line[1],
			MemberEmail: line[2],
		}
		member := models.Member{
			MemberId:    memberdata.MemberID,
			MemberName:  memberdata.MemberName,
			MemberEmail: memberdata.MemberEmail,
		}

		var dbMember models.Member

		db.Where("member_id = ?", member.MemberId).First(&dbMember)
		if dbMember.MemberId != memberdata.MemberID {
			db.Create(&member)
		}
		if dbMember.MemberId == memberdata.MemberID && (dbMember.MemberName != memberdata.MemberName || dbMember.MemberEmail != memberdata.MemberEmail) {
			db.Model(&member).Where("member_id = ?", member.MemberId).Updates(map[string]interface{}{"member_id": memberdata.MemberID, "member_name": memberdata.MemberName, "member_email": memberdata.MemberEmail})
		}
		if dbMember.MemberId == memberdata.MemberID && (dbMember.MemberName != memberdata.MemberName || dbMember.MemberEmail != memberdata.MemberEmail) {
			db.Model(&member).Where("member_id = ?", member.MemberId).Updates(map[string]interface{}{"member_id": memberdata.MemberID, "member_name": memberdata.MemberName, "member_email": memberdata.MemberEmail})
		}
	}
}
