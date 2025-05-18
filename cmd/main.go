package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"log"
	"strconv"
	"strings"
)

type Propiedad struct {
	Valor                  float64
	ValorUF                float64
	NumeroDeHabitaciones   int
	NumeroDeBanos          int
	NumeroDeEstacinamiento int
	SuperficieTotal        int
	SuperficieConstruida   int
	FechaConstruccion      int
	Direccion              string
}

func (p Propiedad) PrintInfo() {
	fmt.Println("Valor:", int(p.Valor))
	fmt.Println("Valor UF:", p.ValorUF)
	fmt.Println("Numero De Habitaciones:", p.NumeroDeHabitaciones)
	fmt.Println("Numero De Banos:", p.NumeroDeBanos)
	fmt.Println("Numero De Estacinamiento:", p.NumeroDeEstacinamiento)
	fmt.Println("Superficie Total:", p.SuperficieTotal)
	fmt.Println("Superficie Construida:", p.SuperficieConstruida)
	fmt.Println("Fecha Construccion:", p.FechaConstruccion)
	fmt.Println("Direccion:", p.Direccion)
}

func CleanMoney(m string) float64 {
	// se limpia el valor, obteniendo solo el numero (ej: 100.000.000)
	strMoney := strings.Fields(m)[1]
	// eliminamos los puntos en el valor, para poder representar el numero en el tipo float
	strMoney = strings.ReplaceAll(strMoney, ".", "")
	strMoney = strings.ReplaceAll(strMoney, ",", ".")

	// hacemos la conversion de string a float
	money, err := strconv.ParseFloat(strMoney, 64)
	if err != nil {
		log.Fatal(err)
	}

	return money
}

func CleanIntegers(n string) int {
	n = strings.Fields(n)[0] // esto se hace para eliminar los prefijos m2 de las superficies
	n = strings.ReplaceAll(n, ".", "")

	num, err := strconv.Atoi(n) // hacemos la conversion de string a int
	if err != nil {
		log.Fatal(err)
	}
	return num
}

func main() {
	/*
		Un collector se encarga de gestionar la comunicación red cmo la ejecución de callbacks, que son funciones
		que se ejecutan en respuesta a eventos especificos durante el trabajo de scraping
	*/
	c := colly.NewCollector()
	collectorPropiedades := c.Clone()

	propiedades := []Propiedad{}

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	/*
		En la página Chilepropiedades, todos los elementos <a></a> que enlazan a propiedades en venta comparten una misma clase.
		Por esta razón, se descartan los demás enlaces, ya que no apuntan a propiedades y se consideran irrelevantes.
		El callback OnHTML se ejecuta cada vez que se encuentra un nodo HTML que coincide con el selector CSS (en este caso a[href]).
	*/
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// filtramos los enlaces que no tienen la clase especifica de propiedades en venta
		if e.Attr("class") != "d-block text-ellipsis clp-big-value" {
			return
		}

		// obtenemos el enlace asociado a una propiedad
		link := e.Attr("href")

		// Realizamos una solicitud para visitar la URL de la propiedad
		collectorPropiedades.Visit(e.Request.AbsoluteURL(link))
	})

	collectorPropiedades.OnRequest(func(r *colly.Request) {
		fmt.Println("Visitando propiedad:", r.URL)
	})

	// Cada vez que encuentre un elemento que contenga la clase "clp-details-table" se ejecutara esta funcion(callback)
	collectorPropiedades.OnHTML(".clp-details-table", func(e *colly.HTMLElement) {
		var currentLabel string
		currentProperty := Propiedad{}

		e.DOM.Children().Each(func(i int, s *goquery.Selection) {
			if s.HasClass("clp-description-label") {
				currentLabel = s.Text()
			} else if s.HasClass("clp-description-value") {
				/* Puedes limpiar texto aquí si es necesario
				value := s.Text()
				if currentLabel == "Valor (CLP aprox.)*:" || currentLabel == "Valor (UF aprox.)*:" || currentLabel == "Valor:" {
					value = CleanMoney(value)
				}
				data[currentLabel] = value*/
				value := s.Text()
				switch currentLabel {
				case "Valor:":
					if strings.HasPrefix(value, "$") {
						currentProperty.Valor = CleanMoney(value)
					} else if strings.HasPrefix(value, "UF") {
						currentProperty.ValorUF = CleanMoney(value)
					}
				case "Valor (UF aprox.)*:":
					currentProperty.ValorUF = CleanMoney(value)
				case "Valor (CLP aprox.)*:":
					currentProperty.Valor = CleanMoney(value)
				case "Habitaciones:":
					currentProperty.NumeroDeHabitaciones = CleanIntegers(value)
				case "Baño:":
					currentProperty.NumeroDeBanos = CleanIntegers(value)
				case "Estacionamientos:":
					currentProperty.NumeroDeEstacinamiento = CleanIntegers(value)
				case "Superficie Total:":
					currentProperty.SuperficieTotal = CleanIntegers(value)
				case "Superficie Construida:":
					currentProperty.SuperficieConstruida = CleanIntegers(value)
				case "Año Construcción:":
					currentProperty.FechaConstruccion = CleanIntegers(value)
				case "Dirección:":
					currentProperty.Direccion = value

				}
			}
		})

		propiedades = append(propiedades, currentProperty)
	})

	c.Visit("https://chilepropiedades.cl/propiedades/venta/casa/region-metropolitana-de-santiago-rm/0")

	for _, p := range propiedades {
		fmt.Println("-------------------------------------------")
		p.PrintInfo()
	}
	fmt.Println("-------------------------------------------")

}
