package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"strings"
    "strconv"
    "reflect"

	rtree "github.com/dhconnelly/rtreego"
	"github.com/rustyoz/svg"
)

func (f Feature) Bounds() *rtree.Rect {
	minX := math.MaxFloat64
	minY := math.MaxFloat64
	maxX := math.SmallestNonzeroFloat64
	maxY := math.SmallestNonzeroFloat64
	for _, segment := range f.Geometry {
		for _, point := range segment.Points {
			x := point[0]
			y := point[1]
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
		}
	}
	point := []float64{minX, minY}
	width := math.Max(maxX-minX, 1)
	height := math.Max(maxY-minY, 1)
	lengths := []float64{width, height}
	rect, err := rtree.NewRect(point, lengths)
	if err != nil {
		fmt.Println("maxX", maxX)
		fmt.Println("maxY", maxY)
		fmt.Println("minX", minX)
		fmt.Println("minY", minY)
		fmt.Println("I must've fucked up the dimension", err, point, lengths)
	}
	return rect
}

type Feature struct {
	Id       string
	D        string
	Geometry []svg.Segment
}

func insertPath(rt *rtree.Rtree, p *svg.Path) {
	segments := p.Parse() // Do this to start sending Segements down path's Segment channel
	feat := Feature{p.Id, p.D, nil}
	for seg := range segments {
		feat.Geometry = append(feat.Geometry, seg)
	}
	rt.Insert(feat)
}

func insertLine(rt *rtree.Rtree, l *svg.Line) {
    x1, _ := strconv.ParseFloat(l.X1, 64)
    y1, _ := strconv.ParseFloat(l.Y1, 64)
    x2, _ := strconv.ParseFloat(l.X2, 64)
    y2, _ := strconv.ParseFloat(l.Y2, 64)
    segment := svg.Segment { 1, true, [][2]float64{[2]float64{x1, y1}, [2]float64{x2, y2}}}
    feat := Feature{l.Id, "", []svg.Segment{segment}}
    rt.Insert(feat)
}

func insertGroup(rt *rtree.Rtree, g *svg.Group) {
	if g.Parent != nil {
		g.Transform.MultiplyWith(*g.Parent.Transform)
	} else {
		g.Transform.MultiplyWith(*g.Owner.Transform)
	}
	for _, elem := range g.Elements {
		switch elem.(type) {
		case *svg.Group:
            fmt.Println("test")
			group := elem.(*svg.Group)
			insertGroup(rt, group)
		case *svg.Path:
            fmt.Println("test")
			insertPath(rt, elem.(*svg.Path))
        case *svg.Line:
            fmt.Println("test")
            insertLine(rt, elem.(*svg.Line))
		default:
            fmt.Println(reflect.Indirect(reflect.ValueOf(&elem)).Elem().Type())
		}
	}
}

func insertElements(rt *rtree.Rtree, svg *svg.Svg) {
	for _, g := range svg.Groups {
		insertGroup(rt, &g)
	}
}

func main() {
	buffer, err := ioutil.ReadFile("test.svg")
	if err != nil {
		fmt.Println("Couldn't read file", err)
		return
	}
	svga, err := svg.ParseSvg(string(buffer), "test.svg", 1.0)
	if err != nil {
		fmt.Println("Couldn't parse svg", err)
		return
	}
	rt := rtree.NewTree(2, 25, 50)
	fmt.Println("Inserting elements")
	insertElements(rt, svga)
	fmt.Println("done, searching")
	bounds, _ := rtree.NewRect([]float64{0, 0}, []float64{500,500})
	results := rt.SearchIntersect(bounds)
	fmt.Println("done")
	fmt.Println(len(results))
	for _, feat := range results {
		geom := feat.(Feature).Geometry
		var paths []string
		for _, seg := range geom {
			for _, point := range seg.Points {
				paths = append(paths, fmt.Sprintf("%v,%v", point[0], point[1]))
			}
		}
		fmt.Printf("<path d=\"M%s\"></path>\n", strings.Join(paths, "L"))
	}
}
