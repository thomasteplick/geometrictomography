/*
Computed Tomography displays the internal structure of 3D geometric objects
such as ellipsoids, parabloids, cubes, planes, or cones.  It slices the
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
	patternCT              = "/computedCT"         // http handler computed tomography
	xlabelsZoom            = 11                    // # labels on x axis in zoom
	ylabelsZoom            = 11                    // # labels on y axis in zoom
	xlabelsOverview        = 1                     // # labels on x axis in overview
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
	Grid   []string // plotting grid
	Status string   // status of the plot
	Xlabel []string // x-axis labels
	Ylabel []string // y-axis labels
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

	densities := make([][][]byte, planeDim)
	for i := range densities {
		densities[i] = make([][]byte, planeDim)
		for j := range densities[i] {
			densities[i][j] = make([]byte, planeDim)
		}
	}

	// Get the planeStarti, planeStartj, planeStartk, planeStep from html form
	txt := r.FormValue("planestarti")
	planeStarti, err := strconv.Atoi(txt)
	if err != nil {
		fmt.Printf("Atoi for planeStarti error: %v\n", err.Error())
		return nil, fmt.Errorf("enter plane start axis i")
	}

	txt = r.FormValue("planestartj")
	planeStartj, err := strconv.Atoi(txt)
	if err != nil {
		fmt.Printf("Atoi for planeStartj error: %v\n", err.Error())
		return nil, fmt.Errorf("enter plane start axis j")
	}

	txt = r.FormValue("planestartk")
	planeStartk, err := strconv.Atoi(txt)
	if err != nil {
		fmt.Printf("Atoi for planeStartk error: %v\n", err.Error())
		return nil, fmt.Errorf("enter plane start axis k")
	}

	txt = r.FormValue("planestep")
	planeStep, err := strconv.Atoi(txt)
	if err != nil {
		fmt.Printf("Atoi for planeStep error: %v\n", err.Error())
		return nil, fmt.Errorf("enter plane step")
	}

	// Read the geometric object file containing the densities
	for i := range planeDim {
		for j := range planeDim {
			for k := range planeDim {
				_, err := fmt.Fscanf(f, "%d", &densities[i][j][k])
				if err != nil {
					fmt.Printf("Fscanf for densities[%d][%d][%d] error: %v\n", i, j, k, err.Error())
					return nil, fmt.Errorf("function Fscanf for densities[%d][%d][%d] error: %v", i, j, k, err.Error())
				}
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
		xmax: planeDim - 1,
		ymin: 0,
		ymax: planeDim - 1,
	}

	return &ct, nil
}

// draw horizontal and vertical lines for Overview
func (ct *ComputedTomography) drawCrossedLines() {
	// draw crossed horizontal and vertical lines for overview
	// loop over 5 horizontal lines, make axis separation double thick
	for y := 0; y < rows; y += planeDim {
		for x := range cols {
			row := rows - y
			col := x
			ct.plot.Grid[row*cols+col] = "online"
		}
	}
	// draw separation between axes double thick
	y := 100
	for x := range cols {
		row := rows - y
		col := x
		ct.plot.Grid[row*cols+col] = "online"
		row = rows - 2*y
		ct.plot.Grid[row*cols+col] = "online"
	}

	// loop over 5 vertical
	for x := 0; x < cols; x += planeDim {
		for y = range rows {
			row := rows - y
			col := x
			ct.plot.Grid[row*cols+col] = "online"
		}
	}
}

// Axial planes overview
func (ct *ComputedTomography) gridFillOverview(r *http.Request) error {
	ct.xmin = 0
	ct.xmax = cols
	ct.ymin = 0.0
	ct.ymax = rows

	txt := r.FormValue("planestep")
	if len(txt) == 0 {
		fmt.Printf("plane step length = 0")
		return fmt.Errorf("enter plane step")
	}
	planeStep, err := strconv.Atoi(txt)
	if err != nil {
		fmt.Printf("plane step int conversion error: %v\n", err.Error())
		return fmt.Errorf("plane step int conversion error: %v", err.Error())
	}
	/**************** axis i *******************/
	planeCnt := 0
	txt = r.FormValue("planestarti")
	if len(txt) == 0 {
		fmt.Printf("plane start axis i length = 0\n")
		return fmt.Errorf("enter plane start for axis i")
	}
	planeStarti, err := strconv.Atoi(txt)
	if err != nil {
		fmt.Printf("plane start i int conversion error: %v\n", err.Error())
		return fmt.Errorf("plane start i int conversion error: %v", err.Error())
	}
	planeStopi := min(planeDim, nplanes*planeStep+planeStarti)

	// loop over the planes
	for i := planeStarti; i < planeStopi; i += planeStep {
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
	txt = r.FormValue("planestartj")
	if len(txt) == 0 {
		fmt.Printf("plane start axis j length = 0\n")
		return fmt.Errorf("enter plane start for axis j")
	}
	planeStartj, err := strconv.Atoi(txt)
	if err != nil {
		fmt.Printf("plane start j int conversion error: %v\n", err.Error())
		return fmt.Errorf("plane start j int conversion error: %v", err.Error())
	}
	planeStopj := min(planeDim, nplanes*planeStep+planeStartj)

	// loop over the planes
	for j := planeStartj; j < planeStopj; j += planeStep {
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
	txt = r.FormValue("planestartk")
	if len(txt) == 0 {
		fmt.Printf("plane start axis k length = 0\n")
		return fmt.Errorf("enter plane start for axis k")
	}
	planeStartk, err := strconv.Atoi(txt)
	if err != nil {
		fmt.Printf("plane start k int conversion error: %v\n", err.Error())
		return fmt.Errorf("plane start k int conversion error: %v", err.Error())
	}
	planeStopk := min(planeDim, nplanes*planeStep+planeStartk)
	// loop over the planes
	for k := planeStartk; k < planeStopk; k += planeStep {
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

	// Construct the y-axis labels
	ct.plot.Ylabel = make([]string, ylabelsZoom)
	y := []string{"C", "B", "A"}
	for i := range ct.plot.Ylabel {
		ct.plot.Ylabel[i] = y[i]
	}

	// Construct the x-axis labels
	return nil
}

// Show sequences of axial planes of the geometric object
func (ct *ComputedTomography) processOverview(r *http.Request) error {
	ct.plot.Grid = make([]string, rows*cols)
	ct.plot.Xlabel = make([]string, xlabelsZoom)
	ct.plot.Ylabel = make([]string, ylabelsZoom)

	// Put axial planes in PlotT grid
	err := ct.gridFillOverview(r)
	if err != nil {
		return fmt.Errorf("gridFillOverview() error: %v", err)
	}

	// Construct the y-axis labels
	ct.plot.Ylabel = make([]string, ylabelsOverview)
	y := []string{"k", "j", "i"}
	for i := range ct.plot.Ylabel {
		ct.plot.Ylabel[i] = y[i]
	}

	// Construct the x-axis labels
	x := "Axial Planes"
	ct.plot.Xlabel[0] = x
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
		plot.Status = fmt.Sprintf("Open file %s error: %v", geometricobject, err.Error())
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
		err := ct.processOverview(r)
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
