package aoi

import (
	"github.com/fogleman/gg"
	"github.com/labstack/gommon/log"
	"math/rand"
	"rocommon/util"
	"strconv"
	"time"
)

var dc = gg.NewContext(800, 800)

type QuadObj struct {
	Obj  interface{}
	Rect *QuadBounds
}
type QuadBounds struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}
type QuadTreeNode struct {
	MaxObjNum int32 //split之前最大能存储的节点数量
	MaxLevels int32 //最大深度
	Level     int32 //当前深度

	Rect    *QuadBounds     //当前节点的bound
	ObjList []*QuadObj      //当前节点存储的对象数量
	Nodes   []*QuadTreeNode //子节点
}

func newQuadTreeNode(level int32, maxLevel int32, rect *QuadBounds) *QuadTreeNode {
	node := &QuadTreeNode{
		MaxLevels: maxLevel,
		Level:     level,
		Rect:      rect,
		MaxObjNum: 2,
	}
	return node
}

func NewQuadTree(maxLevel int32, rect *QuadBounds) *QuadTreeNode {
	root := &QuadTreeNode{
		MaxLevels: maxLevel,
		Rect:      rect,
		MaxObjNum: 2,
	}

	return root
}

func (this *QuadTreeNode) Clear() {
	this.ObjList = []*QuadObj{}
	for idx := 0; idx < len(this.Nodes); idx++ {
		this.Nodes[idx].Clear()
	}
	this.Nodes = []*QuadTreeNode{}
	this.Level = 0
}

//split当前节点成4个子节点
func (this *QuadTreeNode) Split() {
	subWidth := this.Rect.Width * 0.5
	subHeight := this.Rect.Height * 0.5
	x := this.Rect.X
	y := this.Rect.Y

	//-------
	//|2 | 3| top
	//-------
	//|1 | 0| bottom
	//-------
	//^
	//|(y)
	//|
	//|
	//|——————————>(x)
	this.Nodes = append(this.Nodes, newQuadTreeNode(this.Level+1, this.MaxLevels,
		&QuadBounds{X: x + subWidth, Y: y, Width: subWidth, Height: subHeight}))
	this.Nodes = append(this.Nodes, newQuadTreeNode(this.Level+1, this.MaxLevels,
		&QuadBounds{X: x, Y: y, Width: subWidth, Height: subHeight}))
	this.Nodes = append(this.Nodes, newQuadTreeNode(this.Level+1, this.MaxLevels,
		&QuadBounds{X: x, Y: y + subHeight, Width: subWidth, Height: subHeight}))
	this.Nodes = append(this.Nodes, newQuadTreeNode(this.Level+1, this.MaxLevels,
		&QuadBounds{X: x + subWidth, Y: y + subWidth, Width: subWidth, Height: subHeight}))

	dc.SetLineWidth(float64(0.2))
	dc.DrawRectangle(float64(x+subWidth), float64(y), float64(subWidth), float64(subHeight))
	dc.DrawRectangle(float64(x), float64(y), float64(subWidth), float64(subHeight))
	dc.DrawRectangle(float64(x), float64(y+subHeight), float64(subWidth), float64(subHeight))
	dc.DrawRectangle(float64(x+subWidth), float64(y+subHeight), float64(subWidth), float64(subHeight))
	dc.Stroke()
}

//获取包含rect的节点序号0-3
func (this *QuadTreeNode) GetIndexes(rect *QuadBounds) []int32 {
	var indexList []int32
	verticalMidPoint := this.Rect.X + this.Rect.Width*0.5
	horizonMidPoint := this.Rect.Y + this.Rect.Height*0.5

	//判断上下边界
	topQuadrant := rect.Y >= horizonMidPoint
	bottomQuadrant := rect.Y-rect.Height <= horizonMidPoint
	topAndBottomQuadrant := rect.Y+rect.Height >= horizonMidPoint && rect.Y <= horizonMidPoint

	if topAndBottomQuadrant {
		bottomQuadrant = false
		topQuadrant = false
	}
	//判断是否同时在左右两边
	if rect.X+rect.Width >= verticalMidPoint && rect.X <= verticalMidPoint {
		if topQuadrant {
			indexList = append(indexList, 2, 3)
		} else if bottomQuadrant {
			indexList = append(indexList, 0, 1)
		} else if topAndBottomQuadrant {
			indexList = append(indexList, 0, 1, 2, 3)
		}
	} else if rect.X >= verticalMidPoint {
		//判断只在右边
		if topQuadrant {
			indexList = append(indexList, 3)
		} else if bottomQuadrant {
			indexList = append(indexList, 0)
		} else if topAndBottomQuadrant {
			indexList = append(indexList, 0, 3)
		}

	} else if rect.X+rect.Width <= verticalMidPoint {
		//判断只在左边
		if topQuadrant {
			indexList = append(indexList, 2)
		} else if bottomQuadrant {
			indexList = append(indexList, 1)
		} else if topAndBottomQuadrant {
			indexList = append(indexList, 1, 2)
		}
	}
	return indexList
}

func (this *QuadTreeNode) Insert(obj *QuadObj) {
	if len(this.Nodes) > 0 {
		indexList := this.GetIndexes(obj.Rect)
		if len(indexList) > 0 {
			for idx := 0; idx < len(indexList); idx++ {
				this.Nodes[indexList[idx]].Insert(obj)
			}
			return
		}
	}
	this.ObjList = append(this.ObjList, obj)
	if len(this.ObjList) > int(this.MaxObjNum) && this.Level < this.MaxLevels {
		//split
		if len(this.Nodes) <= 0 {
			this.Split()
		}

		idx := 0
		for idx < len(this.ObjList) {
			indexList := this.GetIndexes(this.ObjList[idx].Rect)
			if len(indexList) > 0 {
				for k := 0; k < len(indexList); k++ {
					this.Nodes[indexList[k]].Insert(this.ObjList[idx])
				}
				this.ObjList = append(this.ObjList[:idx], this.ObjList[idx+1:]...)
			} else {
				idx++
			}
		}
	}
}

func (this *QuadTreeNode) Retrieve(obj *QuadObj, ObjList *[]*QuadObj) {
	indexList := this.GetIndexes(obj.Rect)

	*ObjList = append(*ObjList, this.ObjList...)
	if len(this.Nodes) > 0 {
		if len(indexList) > 0 {
			for idx := 0; idx < len(indexList); idx++ {
				this.Nodes[indexList[idx]].Retrieve(obj, ObjList)
			}
		} else {
			for idx := 0; idx < len(this.Nodes); idx++ {
				this.Nodes[idx].Retrieve(obj, ObjList)
			}
		}
	}
}

func TestQuadtree() {
	rand.Seed(int64(util.GetTimeMilliseconds()))
	dc.SetRGB(0, 0, 0)
	dc.Clear()

	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()
	ww := 0.2
	dc.SetRGBA(r, g, b, float64(1))
	dc.SetLineWidth(float64(ww))

	quadTest := NewQuadTree(8, &QuadBounds{0, 0, 800, 800})

	var allObjList []*QuadObj
	//nowTime := time.Now()
	var grid float32 = 10.0
	gridh := quadTest.Rect.Width / grid
	gridv := quadTest.Rect.Height / grid
	for idx := 0; idx < 100; idx++ {
		x := float32(rand.Int31n(int32(gridh))) * grid
		y := float32(rand.Int31n(int32(gridv))) * grid
		w := float32(rand.Intn(4)+1) * grid
		h := float32(rand.Intn(4)+1) * grid
		//w := 1
		//h := 1
		obj := &QuadObj{
			Rect: &QuadBounds{x, y, float32(w), float32(h)},
			Obj:  idx + 1,
		}

		//a := rand.Float64()*0.5 + 0.5
		ww := 1
		dc.SetRGBA(1, 1, 1, float64(100))
		dc.SetLineWidth(float64(ww))
		dc.DrawRectangle(float64(x), float64(y), float64(w), float64(h))
		dc.Stroke()

		quadTest.Insert(obj)

		allObjList = append(allObjList, obj)
	}
	//log.Printf("time=%v", time.Now().Sub(nowTime).String())

	for ii := 0; ii < 50; ii++ {
		dc.SetRGB(0, 0, 0)
		dc.Clear()
		var returnObjList []*QuadObj
		//nowTime = time.Now()
		//for idx := 0; idx < len(allObjList); idx++ {
		//	quadTest.Retrieve(allObjList[idx], returnObjList)
		//}

		r = rand.Float64()
		g = rand.Float64()
		b = rand.Float64()
		returnObjList = returnObjList[:0]
		tmpObj := &QuadObj{
			Rect: &QuadBounds{
				X:      float32(r)*200 + float32(ii)*5,
				Y:      float32(g)*200 + float32(ii)*5,
				Width:  400,
				Height: 400,
			}}
		nowTime := time.Now()
		quadTest.Retrieve(tmpObj, &returnObjList)
		log.Printf("time=%v", time.Now().Sub(nowTime).String())
		x := tmpObj.Rect.X
		y := tmpObj.Rect.Y
		w := tmpObj.Rect.Width
		h := tmpObj.Rect.Height
		dc.SetRGBA(r+22, g, b, float64(10))
		dc.SetLineWidth(float64(2))
		dc.DrawRectangle(float64(x), float64(y), float64(w), float64(h))
		dc.Stroke()

		r = rand.Float64()
		g = rand.Float64()
		b = rand.Float64()
		for idx := 0; idx < len(returnObjList); idx++ {
			x := returnObjList[idx].Rect.X
			y := returnObjList[idx].Rect.Y
			w := returnObjList[idx].Rect.Width
			h := returnObjList[idx].Rect.Height
			ww := 1
			dc.SetRGB(r+100, g+100, b+100)
			dc.SetLineWidth(float64(ww))
			dc.DrawRectangle(float64(x), float64(y), float64(w), float64(h))
			dc.Stroke()
		}
		//log.Printf("time1=%v", time.Now().Sub(nowTime).String())
		dc.SavePNG("rect" + strconv.Itoa(ii) + ".png")
	}
}
