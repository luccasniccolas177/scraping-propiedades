package main

import (
	"fmt"
	"github.com/gocolly/colly"
)

type Propiedad struct {
	Valor                  float64
	NumeroDeHabitaciones   int
	NumeroDeBanos          int
	NumeroDeEstacinamiento int
	SuperficieTotal        int
	SuperficieConstruida   int
	FechaConstruccion      int
	Direccion              string
}

func main() {
	/*
		Un collector se encarga de gestionar la comunicación red cmo la ejecución de callbacks, que son funciones
		que se ejecutan en respuesta a eventos especificos durante el trabajo de scraping
	*/
	c := colly.NewCollector()

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
		e.Request.Visit(link)
	})

	c.Visit("https://chilepropiedades.cl/propiedades/venta/casa/region-metropolitana-de-santiago-rm/0")

	fmt.Println("Adios")
}
