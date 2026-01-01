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
	planeDim        = 50                    // plane dimension in cells in i, j, k data space
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
	// Ax + By + Cz = D, => D=A*x1+B*y1+C*z1
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	A := 2
	B := 3
	C := 4
	D := A*x1 + B*y1 + C*z1
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
	// (x)^2/a^2 + (y)^2/b^2 + (z)^2/c^2 = 1
	var black byte = 9
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	a := x1/2 - 3
	a2 := a * a
	b := y1 / 2
	b2 := b * b
	c := z1/2 + 2
	c2 := c * c
	for x := -a; x < a; x++ {
		for y := -b; y < b; y++ {
			tmp := 1.0 - float64(x*x)/float64(a2) - float64(y*y)/float64(b2)
			if tmp >= 0 {
				z := int(math.Sqrt(tmp * float64(c2)))
				geo.density[x+x1][y+y1][z1+z] = black
				geo.density[x+x1][y+y1][z1-z] = black
			}
		}
	}
}

// ellipsoid, solid
func (geo *GeoObject) createEllipsoidSolid() {
	// center (x1,y1,z1)
	// (x)^2/a^2 + (y)^2/b^2 + (z)^2/c^2 = 1
	black := 9.0
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	// assign axis maximums
	a := x1/2 - 4
	b := y1 / 2
	c := z1/2 + 4
	norm := float64(a*a + b*b + c*c)
	var density byte
	// Fill in the ellipsoid while any axis is greater than or equal to zero
	for i := a; i >= 0; i-- {
		a2 := i * i
		for j := b; j >= 0; j-- {
			b2 := j * j
			for k := c; k >= 0; k-- {
				c2 := k * k
				for x := -i; x <= i; x++ {
					for y := -j; y <= j; y++ {
						tmp := 1.0 - float64(x*x)/float64(a2) - float64(y*y)/float64(b2)
						if tmp >= 0 {
							z := int(math.Sqrt(tmp * float64(c2)))
							sumsq := float64(x*x + y*y + z*z)
							// center is the most dense, decreasing as you move away from center
							density = byte(black * (1.0 - math.Sqrt(sumsq/norm)))
							geo.density[x+x1][y+y1][z1+z] = density
							geo.density[x+x1][y+y1][z1-z] = density
						}
					}
				}
			}
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
	a2 := a * a
	b := y1 / 2
	b2 := b * b
	c := z1
	c2 := c * c
	//z1c := z1 - c/2
	z1c := z1 + c/2
	for x := -a; x < a; x++ {
		for y := -b; y < b; y++ {
			z := int(math.Sqrt((float64(x*x)/float64(a2) + float64(y*y)/float64(b2)) * float64(c2)))
			//geo.density[z1c+z][x1+x][y1+y] = black
			geo.density[z1c-z][x1+x][y1+y] = black
		}
	}
}

// elliptic parabaloid, surface
func (geo *GeoObject) createParaboloid() {
	// center (x1,y1,z1)
	// x1^2/a^2 + y^2/b^2 = z/c
	var black byte = 9
	x1 := planeDim / 2
	y1 := planeDim / 2
	z1 := planeDim / 2
	a := x1 / 2
	a2 := a * a
	b := y1 / 2
	b2 := b * b
	c := z1
	z1c := z1 + c/2
	//z1c := z1 - c/2

	for x := -a; x < a; x++ {
		for y := -b; y < b; y++ {
			z := int((float64(x*x)/float64(a2) + float64(y*y)/float64(b2)) * float64(c))
			if z1c >= z {
				geo.density[z1c-z][x+x1][y+y1] = black
			}
			/*
				if z1c+z < planeDim {
					geo.density[z1c+z][x+x1][y+y1] = black
				}
			*/

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
	case "ellipsoidsurface":
		geo.createEllipsoid()
	case "ellipsoidsolid":
		geo.createEllipsoidSolid()
	case "paraboloid":
		geo.createParaboloid()
	case "cone":
		geo.createCone()
	case "box":
		geo.createBox()
	default:
		fmt.Printf("create object unknown case: '%s'\n", geometricObject)
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
