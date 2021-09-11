package components

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

type GeoCode struct {
	gorm.Model
	Name    string
	JaName  string
	Lat     float64
	Lng     float64
	Country string
}

func (g *GeoCode) CreateGeoCode(tx *gorm.DB) error {
	// first of all, detect if g exist
	gg := new(GeoCode)
	tx.Model(&GeoCode{}).Where("name=?", g.Name).First(gg)
	if gg.Name == g.Name {
		return fmt.Errorf("create thing exist")
	}

	if err := tx.Create(g).Error; err != nil {
		return err
	}
	return nil
}

func QueryGeoCodeByName(name string) (*GeoCode, error) {
	log.Println("sqlite Request: ", name)

	g := new(GeoCode)
	SqliteConn.Model(&GeoCode{}).Where("name = ? OR ja_name = ?", name, name).First(g)
	if g.Name == name || g.JaName == name {
		return g, nil
	}
	return nil, fmt.Errorf("couldn't find")
}

func (g *GeoCode) UpdateGeoCode(tx *gorm.DB) error {
	if err := tx.Model(&GeoCode{}).Where("name = ?", g.Name).Updates(GeoCode{Lat: g.Lat, Lng: g.Lng}).Error; err != nil {
		return err
	}
	return nil
}

func (g *GeoCode) DeleteGeoCode(tx *gorm.DB) error {
	if err := tx.Model(&GeoCode{}).Where("name=?", g.Name).Delete(&GeoCode{}).Error; err != nil {
		return err
	}
	return nil
}
