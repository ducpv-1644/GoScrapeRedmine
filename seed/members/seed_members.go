package members

import (
	"encoding/csv"
	"fmt"
	"go-scrape-redmine/config"
	"go-scrape-redmine/models"
	"go-scrape-redmine/seed"
	"os"
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

func (a *Member) SeedMember() {
	db := config.DBConnect()
	csvFile, err := os.Open(os.Getenv("SEED_FILE_PATH"))
	if err != nil {
		fmt.Println("members.csv not found")
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened CSV file")
	defer csvFile.Close()

	csvLines, err := csv.NewReader(csvFile).ReadAll()
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
			fmt.Println(memberdata.MemberID + " " + memberdata.MemberID + " " + memberdata.MemberName)
		}
	}
}
