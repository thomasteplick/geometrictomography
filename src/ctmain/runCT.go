/*
Geometric Tomography displays the internal structure of 3D geometric objects
such as ellipsoids, parabloids, cubes, boxes, planes, or cones.  It slices the
geometric objects along axial planes in the Cartesian coordinate system.
The object can be solids as well as surfaces.
It gives an overview of the planes in i, j, k axes along with the option
of zooming in on a particular axial plane.  It is possible to select and
view particular planes in the geometric object with different step sizes.
*/

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/thomasteplick/geometricCT"
)

const (
	addr                   = "127.0.0.1:8080"      // http server listen address
	fileComputedTomography = "templates/ct.html"   // html for computed tomography
	patternCT              = "/computedtomography" // http handler computed tomography
	xlabelsZoom            = 11                    // # labels on x axis in zoom
	ylabelsZoom            = 11                    // # labels on y axis in zoom
	ylabelsOverview        = 3                     // # labels on y axis in overview
	geometricobject        = "geometricobject.txt" // 3D geometric object file containing the densities, 50x50x50
	dataDir                = "data/"               // directory for player positions
	rows                   = 300                   // rows in canvas
	cols                   = 300                   // columns in canvas
	planeDim               = 50                    // number of cells in a plane in x and y in overview
	nplanes                = 12                    // number of planes per axis (2 rows) in overview
	nplanes2               = nplanes / 2           // number of planes in each row in overview
	axisDim                = 100                   // number of cells in each axis in y direction in overview
)

// Type to contain all the HTML template actions
type PlotT struct {
	Grid            []string // plotting grid
	Status          string   // status of the plot
	Xlabel          []string // x-axis labels
	Ylabel          []string // y-axis labels
	XlabelContainer string   // x-axis labels
	YlabelContainer string   // x-axis labels
	Domain          string   // Overview or Zoom
}

// Type to hold computed tomography state
type ComputedTomography struct {
	density     [][][]byte // geometric object 3D densities
	plot        *PlotT
	planeStarti int // overview starting plane in i axis
	planeStartj int // overview starting plane in j axis
	planeStartk int // overview starting plane in k axis
	planeStep   int // overiew plane step size
	Endpoints
	density2grayscale [10]string
}

// Type to hold the minimum and maximum data values of the MSE in the Learning Curve
type Endpoints struct {
	xmin int
	xmax int
	ymin int
	ymax int
}

// global variables for parse and execution of the html template
var (
	tmplComputedTomography *template.Template
)

// init parses the html template files
func init() {
	tmplComputedTomography = template.Must(template.ParseFiles(fileComputedTomography))
}

// Construct an instance containing state
func newComputedTomography(r *http.Request, plot *PlotT, f *os.File) (*ComputedTomography, error) {

	var (
		planeStep   int
		planeStarti int
		planeStartj int
		planeStartk int
		err         error
	)
	densities := make([][][]byte, planeDim)
	for i := range densities {
		densities[i] = make([][]byte, planeDim)
		for j := range densities[i] {
			densities[i][j] = make([]byte, planeDim)
		}
	}

	// Get the planeStarti, planeStartj, planeStartk, planeStep from html form
	txt := r.FormValue("planestep")
	if len(txt) == 0 {
		// assign defaults to plane parameters
		planeStep = 4
		planeStarti = 0
		planeStartj = 0
		planeStartk = 0

	} else {
		planeStep, err = strconv.Atoi(txt)
		if err != nil {
			fmt.Printf("Atoi for planeStep error: %v\n", err.Error())
			return nil, fmt.Errorf("enter plane step")
		}

		txt = r.FormValue("planestarti")
		planeStarti, err = strconv.Atoi(txt)
		if err != nil {
			fmt.Printf("Atoi for planeStarti error: %v\n", err.Error())
			return nil, fmt.Errorf("enter plane start axis i")
		}

		txt = r.FormValue("planestartj")
		planeStartj, err = strconv.Atoi(txt)
		if err != nil {
			fmt.Printf("Atoi for planeStartj error: %v\n", err.Error())
			return nil, fmt.Errorf("enter plane start axis j")
		}

		txt = r.FormValue("planestartk")
		planeStartk, err = strconv.Atoi(txt)
		if err != nil {
			fmt.Printf("Atoi for planeStartk error: %v\n", err.Error())
			return nil, fmt.Errorf("enter plane start axis k")
		}
	}

	// Read the geometric object file containing the densities
	for i := range planeDim {
		for j := range planeDim {
			for k := range planeDim - 1 {
				_, err := fmt.Fscanf(f, "%d", &densities[i][j][k])
				if err != nil {
					fmt.Printf("Fscanf for densities[%d][%d][%d] error: %v\n", i, j, k, err.Error())
					return nil, fmt.Errorf("function Fscanf for densities[%d][%d][%d] error: %v", i, j, k, err.Error())
				}
			}
			_, err := fmt.Fscanf(f, "%d\n", &densities[i][j][planeDim-1])
			if err != nil {
				fmt.Printf("Fscanf for densities[%d][%d] newline error: %v\n", i, j, err.Error())
				return nil, fmt.Errorf("function Fscanf for densities[%d][%d] newline error: %v", i, j, err.Error())
			}
		}
	}

	ct := ComputedTomography{
		plot:        plot,
		planeStarti: planeStarti,
		planeStartj: planeStartj,
		planeStartk: planeStartk,
		planeStep:   planeStep,
		density:     densities,
	}
	// Create density2grayscale map
	ct.density2grayscale = [10]string{"gs0", "gs1", "gs2", "gs3", "gs4",
		"gs5", "gs6", "gs7", "gs8", "gs9"}

	// Used in zoom, not overview
	ct.Endpoints = Endpoints{
		xmin: 0,
		xmax: planeDim,
		ymin: 0,
		ymax: planeDim,
	}

	return &ct, nil
}

// draw horizontal and vertical lines for Overview
func (ct *ComputedTomography) drawCrossedLines() {
	// draw crossed horizontal and vertical lines for overview
	// loop over 5 horizontal lines, make axis separation double thick
	for y := 0; y < rows; y += planeDim {
		for x := range cols {
			row := y
			col := x
			ct.plot.Grid[row*cols+col] = "online"
		}
	}
	// draw separation between axes double thick
	y := 100
	for x := range cols {
		col := x
		row := y - 1
		ct.plot.Grid[row*cols+col] = "online"
		row = 2*y - 1
		ct.plot.Grid[row*cols+col] = "online"
		row = y + 1
		ct.plot.Grid[row*cols+col] = "online"
		row = 2*y + 1
		ct.plot.Grid[row*cols+col] = "online"
	}

	// loop over 5 vertical
	for x := 0; x < cols; x += planeDim {
		for y = range rows {
			col := x
			row := y
			ct.plot.Grid[row*cols+col] = "online"
		}
	}
}

// Axial planes overview
func (ct *ComputedTomography) gridFillOverview() error {
	ct.xmin = 0
	ct.xmax = cols
	ct.ymin = 0.0
	ct.ymax = rows

	/**************** axis i *******************/
	planeCnt := 0
	planeStopi := min(planeDim, nplanes*ct.planeStep+ct.planeStarti)

	// loop over the planes
	for i := ct.planeStarti; i < planeStopi; i += ct.planeStep {
		ystart := axisDim - (planeCnt/nplanes2)*planeDim
		xstart := (planeCnt % nplanes2) * planeDim
		for j := 0; j < planeDim; j++ {
			y := ystart - j
			for k := 0; k < planeDim; k++ {
				x := xstart + k
				row := ct.ymax - y
				col := x - ct.xmin
				ct.plot.Grid[row*cols+col] = ct.density2grayscale[ct.density[i][j][k]]
			}
		}
		planeCnt++
	}

	/************* axis j *******************/
	planeCnt = 0
	planeStopj := min(planeDim, nplanes*ct.planeStep+ct.planeStartj)

	// loop over the planes
	for j := ct.planeStartj; j < planeStopj; j += ct.planeStep {
		ystart := 2*axisDim - (planeCnt/nplanes2)*planeDim
		xstart := (planeCnt % nplanes2) * planeDim
		for i := 0; i < planeDim; i++ {
			y := ystart - i
			for k := 0; k < planeDim; k++ {
				x := xstart + k
				row := ct.ymax - y
				col := x - ct.xmin
				ct.plot.Grid[row*cols+col] = ct.density2grayscale[ct.density[i][j][k]]
			}
		}
		planeCnt++
	}

	/******************* axis k ***********************/
	planeCnt = 0
	planeStopk := min(planeDim, nplanes*ct.planeStep+ct.planeStartk)
	// loop over the planes
	for k := ct.planeStartk; k < planeStopk; k += ct.planeStep {
		ystart := 3*axisDim - (planeCnt/nplanes2)*planeDim
		xstart := (planeCnt % nplanes2) * planeDim
		for i := 0; i < planeDim; i++ {
			y := ystart - i
			for j := 0; j < planeDim; j++ {
				x := xstart + j
				row := ct.ymax - y
				col := x - ct.xmin
				ct.plot.Grid[row*cols+col] = ct.density2grayscale[ct.density[i][j][k]]
			}
		}
		planeCnt++
	}

	ct.drawCrossedLines()
	return nil
}

// Zoom plot for the given axis and plane
func (ct *ComputedTomography) gridFillZoom(zoomAxis string, zoomPlane int) error {
	const dupl = 6

	// determine which axis i, j, or k
	switch zoomAxis {
	// axis i
	case "i":
		// convert 50x50 plane to 300x300 grid by duplicating 6x in x and y
		for yin := 0; yin < planeDim; yin++ {
			yout := yin * dupl
			for xin := 0; xin < planeDim; xin++ {
				xout := xin * dupl
				// duplicate the input plane 6x in plot.Grid
				for i := 0; i < dupl; i++ {
					row := yout + i
					for j := 0; j < dupl; j++ {
						col := xout + j
						ct.plot.Grid[row*cols+col] = ct.density2grayscale[ct.density[zoomPlane][yin][xin]]
					}
				}
				xout += dupl
			}
		}

	// axis j
	case "j":
		// convert 50x50 plane to 300x300 grid by duplicating 6x in x and y
		for yin := 0; yin < planeDim; yin++ {
			yout := yin * dupl
			for xin := 0; xin < planeDim; xin++ {
				xout := xin * dupl
				// duplicate the input plane 6x in plot.Grid
				for i := 0; i < dupl; i++ {
					row := yout + i
					for j := 0; j < dupl; j++ {
						col := xout + j
						ct.plot.Grid[row*cols+col] = ct.density2grayscale[ct.density[yin][zoomPlane][xin]]
					}
				}
				xout += dupl
			}
		}

	// axis k
	case "k":
		// convert 50x50 plane to 300x300 grid by duplicating 6x in x and y
		for yin := 0; yin < planeDim; yin++ {
			yout := yin * dupl
			for xin := 0; xin < planeDim; xin++ {
				xout := xin * dupl
				// duplicate the input plane 6x in plot.Grid
				for i := 0; i < dupl; i++ {
					row := yout + i
					for j := 0; j < dupl; j++ {
						col := xout + j
						ct.plot.Grid[row*cols+col] = ct.density2grayscale[ct.density[yin][xin][zoomPlane]]
					}
				}
				xout += dupl
			}
		}
	}

	return nil
}

// insertLabels inserts x- an y-axis labels in the plot
func (ct *ComputedTomography) insertLabels() {
	// Construct x-axis labels
	incr := (ct.xmax - ct.xmin) / (xlabelsZoom - 1)
	x := ct.xmin
	for i := range ct.plot.Xlabel {
		ct.plot.Xlabel[i] = strconv.Itoa(x)
		x += incr
	}

	// Construct the y-axis labels
	incr = (ct.ymax - ct.ymin) / (ylabelsZoom - 1)
	y := ct.ymin
	for i := range ct.plot.Ylabel {
		ct.plot.Ylabel[i] = strconv.Itoa(y)
		y += incr
	}
}

// Expand a particular axial plane in the geometric object
func (ct *ComputedTomography) processZoom(zoomAxis string, zoomPlane int) error {
	ct.plot.Grid = make([]string, rows*cols)
	ct.plot.Xlabel = make([]string, xlabelsZoom)
	ct.plot.Ylabel = make([]string, ylabelsZoom)

	// Put expanded axial plane in PlotT grid
	err := ct.gridFillZoom(zoomAxis, zoomPlane)
	if err != nil {
		return fmt.Errorf("gridFillZoom() error: %v", err)
	}

	// insert x-labels and y-labels in PlotT
	ct.insertLabels()

	// different alignment
	ct.plot.YlabelContainer = "ylabel-zoom"
	ct.plot.XlabelContainer = "xlabel-zoom"

	// plot type
	ct.plot.Domain = fmt.Sprintf("Axial Plane Zoom, Axis=%s, Plane=%d", zoomAxis, zoomPlane)

	return nil
}

// Show sequences of axial planes of the geometric object
func (ct *ComputedTomography) processOverview() error {
	ct.plot.Grid = make([]string, rows*cols)
	ct.plot.Xlabel = make([]string, 2)
	ct.plot.Ylabel = make([]string, 3)

	// Put axial planes in PlotT grid
	err := ct.gridFillOverview()
	if err != nil {
		return fmt.Errorf("gridFillOverview() error: %v", err)
	}

	// Construct the y-axis labels, specify the axes
	ct.plot.Ylabel = make([]string, ylabelsOverview)
	y := []string{"i", "j", "k"}
	for i := range ct.plot.Ylabel {
		ct.plot.Ylabel[i] = y[i]
	}

	// different alignment
	ct.plot.YlabelContainer = "ylabel-overview"
	ct.plot.XlabelContainer = "xlabel-overview"

	// Construct the x-axis labels, just note planes
	ct.plot.Xlabel[0] = "Axial"
	ct.plot.Xlabel[1] = "Planes"

	// plot type
	ct.plot.Domain = fmt.Sprintf("Axial Plane Overview, Plane Step = %d, Axis i start = %d, Axis j start = %d, Axis k start = %d",
		ct.planeStep, ct.planeStarti, ct.planeStartj, ct.planeStartk)

	return nil
}

// runs computed tomography on the geometric object
func handleComputedTomography(w http.ResponseWriter, r *http.Request) {

	var (
		plot PlotT
		ct   *ComputedTomography
	)

	// Determine if a new geometric object is wanted
	txt := r.FormValue("newgeometric")
	if len(txt) > 0 {
		geometricObject := r.FormValue("geometricobject")
		err := geometricCT.CreateObject(geometricObject)
		if err != nil {
			fmt.Printf("createObject() error: %v\n", err)
			plot.Status = fmt.Sprintf("createObject() error: %v", err.Error())
			// Write to HTTP using template and grid
			if err := tmplComputedTomography.Execute(w, plot); err != nil {
				log.Fatalf("Write to HTTP output using template with error: %v\n", err)
			}
			return
		}
	}

	// Open the geometric object file containing the densities
	f, err := os.Open(filepath.Join(dataDir, geometricobject))
	if err != nil {
		fmt.Printf("Open file %s error: %v\n", geometricobject, err)
		plot.Status = "Check New Geometric Object and select the geometric object"
		// Write to HTTP using template and grid
		if err := tmplComputedTomography.Execute(w, plot); err != nil {
			log.Fatalf("Write to HTTP output using template with error: %v\n", err)
		}
		return
	}
	defer f.Close()

	// create ComputedTomography instance to hold state
	ct, err = newComputedTomography(r, &plot, f)
	if err != nil {
		fmt.Printf("newComputedTomography() error: %v\n", err)
		plot.Status = fmt.Sprintf("newComputedTomography error: %v", err.Error())
		// Write to HTTP using template and grid
		if err := tmplComputedTomography.Execute(w, plot); err != nil {
			log.Fatalf("Write to HTTP output using template with error: %v\n", err)
		}
		return
	}

	// expand a plane for the geometric object
	txt = r.FormValue("zoom")
	if txt == "zoomaxisplane" {
		zoomAxis := r.FormValue("zoomaxis")
		txt = r.FormValue("zoomplane")
		zoomPlane, err := strconv.Atoi(txt)
		if err != nil {
			fmt.Printf("zoomPlane int conversion error: %v\n", err.Error())
			plot.Status = fmt.Sprintf("zoomPlane int conversion error: %v\n", err.Error())
			// Write to HTTP using template and grid
			if err := tmplComputedTomography.Execute(w, plot); err != nil {
				log.Fatalf("Write to HTTP output using template with error: %v\n", err)
			}
			return
		}
		err = ct.processZoom(zoomAxis, zoomPlane)
		if err != nil {
			fmt.Printf("processZoom error: %v\n", err.Error())
			plot.Status = fmt.Sprintf("processZoom error: %v\n", err.Error())
			// Write to HTTP using template and grid
			if err := tmplComputedTomography.Execute(w, plot); err != nil {
				log.Fatalf("Write to HTTP output using template with error: %v\n", err)
			}
			return
		}
		// show Overview of geometric object density
	} else {
		err := ct.processOverview()
		if err != nil {
			fmt.Printf("processOverview error: %v\n", err.Error())
			plot.Status = fmt.Sprintf("processOverview error: %v\n", err.Error())
			// Write to HTTP using template and grid
			if err := tmplComputedTomography.Execute(w, plot); err != nil {
				log.Fatalf("Write to HTTP output using template with error: %v\n", err)
			}
			return
		}
	}

	// Execute data on HTML template
	// Write to HTTP using template and grid
	if err := tmplComputedTomography.Execute(w, ct.plot); err != nil {
		log.Fatalf("Write to HTTP output using template with error: %v\n", err)
	}
}

// executive creates the HTTP handlers, listens and serves
func main() {
	// Set up HTTP servers with handlers for computed tomography

	// Create HTTP handler for performing CT
	http.HandleFunc(patternCT, handleComputedTomography)
	fmt.Printf("Computed Tomography Server listening on %v.\n", addr)
	http.ListenAndServe(addr, nil)
}
