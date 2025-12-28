/*
Create geometric object densities for paraboloid, ellipsoid,
plane, cone, box, and cube.  These can be surfaces or solids.
*/

package geometricCT

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

const (
	planeDim        = 50                    // plane dimension in cells in i, j, k
	geometricobject = "geometricobject.txt" // 3D geometric object file containing the densities, 50x50x50
	dataDir         = "data/"               // directory for player positions
)

// 3D geometric object
type GeoObject struct {
	density [][][]byte
}

// plane, surface
func (geo *GeoObject) createPlane() {
	// choose center of plane (x1, y1, z1) = (25, 25, 25)
	// vary x, y, in (0, 49)
	// Normal to plane is Ai + Bj + Ck
	// dot product: A(x-x1) + B(y-y1) + C(z-z1) = 0
	// Ax + By + Cz = D, let A=B=C=1, => D=x1+y1+z1
	A := 1
	B := 1
	C := 1
	D := 3 * planeDim
	var black byte = 9

	for x := 0; x < planeDim; x++ {
		for y := 0; y < planeDim; y++ {
			z := (D - A*x - B*y) / C
			// this point is inside the data space and in the plane
			if z >= 0 && z < planeDim {
				geo.density[x][y][z] = black
			}
		}
	}
}

// cube, solid with decreasing density from center
func (geo *GeoObject) createCube() {
	black := 9.0
	// choose center of cube (x1,y1,z1) = (25,25,25)
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	del := planeDim / 4
	norm := 3.0 * float64(del*del)
	for x := x1 - del; x < x1+del; x++ {
		for y := y1 - del; y < y1+del; y++ {
			for z := z1 - del; z < z1+del; z++ {
				delx := float64(x - x1)
				dely := float64(y - y1)
				delz := float64(z - z1)
				sumsq := delx*delx + dely*dely + delz*delz
				// center is the most dense, decreasing as you move away from center
				geo.density[x][y][z] = byte(black * (1.0 - math.Sqrt(sumsq/norm)))
			}
		}
	}
}

// box, surface
func (geo *GeoObject) createBox() {
	var black byte = 9
	// choose center of box (x1,y1,z1) = (25,25,25)
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	del := planeDim / 4
	for x := x1 - del; x < x1+del; x++ {
		for y := y1 - del; y < y1+del; y++ {
			geo.density[x][y][z1-del] = black
			geo.density[x][y][z1+del] = black
		}
	}
	for y := y1 - del; y < y1+del; y++ {
		for z := z1 - del; z < z1+del; z++ {
			geo.density[x1-del][y][z] = black
			geo.density[x1+del][y][z] = black
		}
	}
	for x := x1 - del; x < x1+del; x++ {
		for z := z1 - del; z < z1+del; z++ {
			geo.density[x][y1-del][z] = black
			geo.density[x][y1+del][z] = black
		}
	}
}

// ellipsoid, surface
func (geo *GeoObject) createEllipsoid() {
	// center (x1,y1,z1)
	// (x-x1)^2/a^2 + (y-y1)^2/b^2 + (z-z1)^2/c^2 = 1
	var black byte = 9
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	a := x1/2 - 3
	b := y1 / 2
	c := z1/2 + 2
	for x := x1 - a; x < x1+a; x++ {
		for y := y1 - b; y < y1+b; y++ {
			z := int(math.Sqrt((1.0-float64(x*x)/float64(a*a)-float64(y*y)/float64(b*b))*float64(c*c))) + z1
			geo.density[x][y][z] = black
		}
	}
}

// elliptic cone, surface
func (geo *GeoObject) createCone() {
	// center (x1,y1,z1)
	// (x)^2/a^2 + (y)^2/b^2 = (z)^2/c^2
	var black byte = 9
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	a := x1 / 2
	b := y1 / 2
	c := z1
	//z1c := z1 - c/2
	z1c := z1 + c/2
	for x := -a; x < a; x++ {
		for y := -b; y < b; y++ {
			z := int(math.Sqrt((float64(x*x)/float64(a*a) + float64(y*y)/float64(b*b)) * float64(c*c)))
			//geo.density[z1c+z][x1+x][y1+y] = black
			geo.density[z1c-z][x1+x][y1+y] = black
		}
	}
}

// elliptic parabaloid, surface
func (geo *GeoObject) createParaboloid() {
	// center (x1,y1,z1)
	// (x-x1)^2/a^2 + (y-y1)^2/b^2 = (z-z1)/c
	var black byte = 9
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	a := x1/2 - 2
	b := y1 / -4
	c := z1/2 + 3
	for x := x1 - a; x < x1+a; x++ {
		for y := y1 - b; y < y1+b; y++ {
			z := c*int(float64(x*x)/float64(a*a)+float64(y*y)/float64(b*b)) + z1
			geo.density[x][y][z] = black
		}
	}

}

// create a geometric 3D object using its densities
func CreateObject(geometricObject string) error {
	// create a GeoObject instance
	geo := GeoObject{density: make([][][]byte, planeDim)}
	for i := range geo.density {
		geo.density[i] = make([][]byte, planeDim)
		for j := range geo.density[i] {
			geo.density[i][j] = make([]byte, planeDim)
		}
	}

	// determine the geometric surface/solid
	switch geometricObject {
	case "plane":
		geo.createPlane()
	case "cube":
		geo.createCube()
	case "ellipsoid":
		geo.createEllipsoid()
	case "paraboloid":
		geo.createParaboloid()
	case "cone":
		geo.createCone()
	case "box":
		geo.createBox()
	default:
		fmt.Printf("Unknown case %s\n", geometricObject)
		return fmt.Errorf("Unknown case %s", geometricObject)
	}

	// Save geometric object
	f, err := os.Create(filepath.Join(dataDir, geometricobject))
	if err != nil {
		fmt.Printf("create %s error: %v\n", geometricObject, err.Error())
		return fmt.Errorf("create %s error: %v", geometricObject, err.Error())
	}
	defer f.Close()

	for i := 0; i < planeDim; i++ {
		for j := 0; j < planeDim; j++ {
			for k := 0; k < planeDim; k++ {
				fmt.Fprintf(f, "%d ", geo.density[i][j][k])
			}
			fmt.Fprintln(f)
		}
	}
	return nil
}
