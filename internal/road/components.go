package road

import "map_generator/internal/terrain"

type Components struct {
	Labels []int
	W, H   int
}

func BuildComponents(v terrain.View) Components {
	w, h := v.Width(), v.Height()
	labels := make([]int, w*h)
	for i := range labels {
		labels[i] = -2
	}
	next := 0

	dirs := [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			if labels[idx] != -2 {
				continue
			}
			if v.IsWater(x, y) {
				labels[idx] = -1
				continue
			}

			queue := []int{idx}
			labels[idx] = next
			for len(queue) > 0 {
				cur := queue[0]
				queue = queue[1:]
				cx, cy := cur%w, cur/w
				for _, d := range dirs {
					nx, ny := cx+d[0], cy+d[1]
					if nx < 0 || nx >= w || ny < 0 || ny >= h {
						continue
					}
					ni := ny*w + nx
					if labels[ni] != -2 {
						continue
					}
					if v.IsWater(nx, ny) {
						labels[ni] = -1
						continue
					}
					labels[ni] = next
					queue = append(queue, ni)
				}
			}
			next++
		}
	}
	return Components{Labels: labels, W: w, H: h}
}

func (c Components) At(x, y int) int { return c.Labels[y*c.W+x] }
