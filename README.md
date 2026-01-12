<h3> 
Geometric Tomography for Three Dimensional Surfaces and Solids
</h3>
<p>
This is a web application written in Go that make use of the html/template package to dynamically
create the web page. Start the web server at bin\ctmain.exe and connect to it from your web browser
at http://127.0.0.1:8080/computedtomography. Each Cartesian axis (i,j,k) of the geometric object consists
of 50 planes, each plane has 50x50 cells.  The cell is a density or concentration which is displayed in grayscale.
There are ten grayscale levels with black representing the greatest density and white being the least dense.
The density or concentration can represent any quantity, not just mass/volume.  The centers of the solid objects are the
most dense, with density decreasing when moving outward with the distance from the center. The geometric 
object can be rotated about any of the (i,j,k) axes for angles(+/-) 180 deg. In the plane overview, the starting
plane for each axis and the plane step size can be chosen to view different parts of the geometric object.
</p>
<h4>Plane Tomogram Overview of Ellipsoid solid</h4>
<h4>Plane Zoom of Ellipsoid solid for dimension i and plane 25</h4>
<h4>Plane Tomogram Overview of Paraboloid surface</h4>
<h4>Plane Zoom of Paraboloid surface for dimension k and plane 25</h4>
<h4>Plane Rotation of 45 degrees for Cone surface</h4>
<h4>Plane Tomogram Overview of Hyperbolic Paraboloid surface</h4>
