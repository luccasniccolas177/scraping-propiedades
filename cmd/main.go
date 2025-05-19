package main

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"log"
	"os"
	"strconv"
	"strings"
)

// Propiedad representa los datos de una propiedad inmobiliaria (casa/depto.) que incluye diferentes
// caracteristicas extraidas desde chlepropiedades.cl
type Propiedad struct {
	Valor                    float64
	ValorUF                  float64
	NumeroDeHabitaciones     int
	NumeroDeBanos            int
	NumeroDeEstacionamientos int
	SuperficieTotal          float64
	SuperficieConstruida     float64
	FechaConstruccion        int
	Comuna                   string
	Direccion                string
	URL                      string
	TipoVivienda             string
	QuienVende               string
	Corredor                 string
}

func (p Propiedad) PrintInfo() {
	fmt.Println("Valor:", int(p.Valor))
	fmt.Println("Valor UF:", p.ValorUF)
	fmt.Println("Numero De Habitaciones:", p.NumeroDeHabitaciones)
	fmt.Println("Numero De Banos:", p.NumeroDeBanos)
	fmt.Println("Numero De Estacionamientos:", p.NumeroDeEstacionamientos)
	fmt.Println("Superficie Total:", p.SuperficieTotal)
	fmt.Println("Superficie Construida:", p.SuperficieConstruida)
	fmt.Println("Fecha Construccion:", p.FechaConstruccion)
	fmt.Println("Direccion:", p.Direccion)
	fmt.Println("Comuna:", p.Comuna)
	fmt.Println("URL:", p.URL)
	fmt.Println("Tipo Vivienda:", p.TipoVivienda)
	fmt.Println("Corredora:", p.Corredor)
	fmt.Println("Quien Vende:", p.QuienVende)
}

// CleanMoney es una función auxiliar que limpia valores monetarios y los representa valor
// flotante en caso de contar con decimales
func CleanMoney(m string) float64 {
	// se limpia el valor, obteniendo solo el numero (ej: 100.000.000)
	strMoney := strings.Fields(m)[1]
	// eliminamos los puntos en el valor, para poder representar el número en el tipo float
	strMoney = strings.ReplaceAll(strMoney, ".", "")
	strMoney = strings.ReplaceAll(strMoney, ",", ".")

	// hacemos la conversion de string a float
	money, err := strconv.ParseFloat(strMoney, 64)
	if err != nil {
		log.Fatal(err)
	}

	return money
}

// CleanIntegers es una función auxiliar para limpiar valores enteros
func CleanIntegers(n string) int {
	n = strings.Fields(n)[0] // esto se hace para eliminar los prefijos m2 de las superficies
	n = strings.ReplaceAll(n, ".", "")

	num, err := strconv.Atoi(n) // hacemos la conversion de string a int
	if err != nil {
		log.Fatal(err)
	}
	return num
}

// CleanArea es una función auxiliar para limpiar los valores de las áreas, borrar el m2
// y transformar las comas a puntos para hacer la conversión correcta
func CleanArea(a string) float64 {
	a = strings.Fields(a)[0]
	a = strings.ReplaceAll(a, ".", "")
	a = strings.ReplaceAll(a, ",", ".")

	area, err := strconv.ParseFloat(a, 64)
	if err != nil {
		log.Fatal(err)
	}

	return area
}

func main() {
	c := colly.NewCollector()
	collectorPropiedades := c.Clone()

	var propiedades []Propiedad
	visitedProperties := make(map[string]bool)

	// Recolectar propiedades, para evitar visitar url's previamente visitadas
	collectorPropiedades.OnRequest(func(r *colly.Request) {
		if visitedProperties[r.URL.String()] {
			r.Abort()
			return
		}
		visitedProperties[r.URL.String()] = true
	})

	collectorPropiedades.OnHTML("body", func(e *colly.HTMLElement) {
		var currentLabel string
		currentProperty := Propiedad{}
		currentProperty.URL = e.Request.URL.String()

		e.DOM.Find(".clp-details-table").Each(func(i int, s *goquery.Selection) {
			s.Children().Each(func(i int, s *goquery.Selection) { // <- Aquí está el cambio correcto
				if s.HasClass("clp-description-label") {
					currentLabel = strings.TrimSpace(s.Text())
				} else if s.HasClass("clp-description-value") {
					value := strings.TrimSpace(s.Text())
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
					case "Baño:", "Baños:":
						currentProperty.NumeroDeBanos = CleanIntegers(value)
					case "Estacionamientos:":
						currentProperty.NumeroDeEstacionamientos = CleanIntegers(value)
					case "Superficie Total:":
						currentProperty.SuperficieTotal = CleanArea(value)
					case "Superficie Construida:":
						currentProperty.SuperficieConstruida = CleanArea(value)
					case "Año Construcción:":
						currentProperty.FechaConstruccion = CleanIntegers(value)
					case "Dirección:":
						direccion := strings.Split(value, ",")
						currentProperty.Comuna = strings.TrimSpace(direccion[0])
						if len(direccion) > 1 {
							currentProperty.Direccion = strings.TrimSpace(strings.Join(direccion[1:], ","))
						}
					case "Tipo de propiedad:":
						currentProperty.TipoVivienda = value
					}
				}
			})
		})

		// Extraer información sobre quien vende y si hay corredora
		e.DOM.Find(".clp-publication-contact-box").Each(func(_ int, box *goquery.Selection) {
			box.Find("h2.subtitle").Each(func(_ int, h2 *goquery.Selection) {
				h2Text := strings.TrimSpace(strings.ToLower(h2.Text()))
				table := h2.Next() // usamos Next() en vez de NextFiltered para mayor compatibilidad

				if h2Text == "información de contacto" {
					// Nombre de quien vende
					nombre := strings.TrimSpace(table.Find("tr").First().Find("td").Text())
					if nombre != "" {
						currentProperty.QuienVende = nombre
					}
				} else if h2Text == "corredora" {
					// Nombre de la corredora
					corredora := strings.TrimSpace(table.Find("tr").First().Find("td").Text())
					if corredora != "" {
						currentProperty.Corredor = corredora
					}
				}
			})

			// Si no se encontró una corredora, asumimos que es venta directa
			if currentProperty.Corredor == "" {
				currentProperty.Corredor = "Dueño Directo"
			}
		})

		propiedades = append(propiedades, currentProperty)
	})

	/*
		En la página Chilepropiedades, todos los elementos <a></a> que enlazan a propiedades en venta comparten una misma clase.
		Por esta razón, se descartan los demás enlaces, ya que no apuntan a propiedades y se consideran irrelevantes.
		El callback OnHTML se ejecuta cada vez que se encuentra un nodo HTML que coincide con el selector CSS (en este caso a[href]).
	*/
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if e.Attr("class") == "d-block text-ellipsis clp-big-value" {
			link := e.Request.AbsoluteURL(e.Attr("href"))
			collectorPropiedades.Visit(link)
		}
	})

	// Buscar la última página y recorrer todas desde 0 hasta la última
	lastPage := 0
	c.OnHTML(".pagination.d-none.d-sm-flex", func(e *colly.HTMLElement) {
		e.DOM.Find("a").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			href, exists := s.Attr("href")
			if exists && strings.Contains(text, "Última") {
				parts := strings.Split(href, "/")
				pageStr := parts[len(parts)-1]
				page, err := strconv.Atoi(pageStr)
				if err == nil && page > lastPage {
					lastPage = page
				}
			}
		})
	})

	// Visitar la primera página (para obtener lastPage)
	startURL := "https://chilepropiedades.cl/propiedades/venta/casa/region-metropolitana-de-santiago-rm/0"
	c.Visit(startURL)

	// Esperamos a que termine para saber la última página
	c.Wait()

	fmt.Printf("Total de páginas: %d\n", lastPage)

	// Visitar todas las páginas
	for i := 0; i <= lastPage; i++ {
		url := fmt.Sprintf("https://chilepropiedades.cl/propiedades/venta/casa/region-metropolitana-de-santiago-rm/%d", i)
		fmt.Println("Visitando página:", url)
		c.Visit(url)
	}

	// Esperamos a que terminen todas las visitas
	c.Wait()
	collectorPropiedades.Wait()

	// Imprimir resultados
	/*for _, p := range propiedades {
		fmt.Println("-------------------------------------------")
		p.PrintInfo()
	}
	fmt.Println("-------------------------------------------")*/

	// Crear archivo CSV con encabezados
	file, err := os.Create("propiedades.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Escribir BOM para que Excel reconozca UTF-8
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	// Crear writer y establecer punto y coma como delimitador para Excel
	w := csv.NewWriter(file)
	w.Comma = ';' // <- Esto corrige el problema en Excel
	defer w.Flush()

	// Escribir encabezados
	headers := []string{
		"Comuna", "Link", "Tipo_Vivienda", "N_Habitaciones", "N_Baños", "N_Estacionamientos",
		"Total_Superficie", "Superficie_Construida", "Valor_UF", "Valor_CLP",
		"Dirección", "Quien_Vende", "Corredor",
	}
	if err := w.Write(headers); err != nil {
		log.Fatal("error writing headers to file", err)
	}

	// Escribir filas
	for _, p := range propiedades {
		row := []string{
			p.Comuna,
			p.URL,
			p.TipoVivienda,
			strconv.Itoa(p.NumeroDeHabitaciones),
			strconv.Itoa(p.NumeroDeBanos),
			strconv.Itoa(p.NumeroDeEstacionamientos),
			fmt.Sprintf("%.2f", p.SuperficieTotal),
			fmt.Sprintf("%.2f", p.SuperficieConstruida),
			fmt.Sprintf("%.2f", p.ValorUF),
			fmt.Sprintf("%.0f", p.Valor),
			p.Direccion,
			p.QuienVende,
			p.Corredor,
		}

		if err := w.Write(row); err != nil {
			log.Fatal("error writing record to file", err)
		}
	}
}
