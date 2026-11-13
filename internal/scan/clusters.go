package scan

import (
	"sort"

	"github.com/cansarihan/mimic/internal/fingerprint"
	"github.com/cansarihan/mimic/internal/srcwalk"
)

// ClusterMember is one function in a clone cluster.
type ClusterMember struct {
	Func string `json:"func"`
	File string `json:"file"`
	Line int    `json:"line"`
}

// Cluster is a group of functions that are structurally alike — a copy-paste
// family. A bug in one is probably in all of them.
type Cluster struct {
	ID      int             `json:"id"`
	Size    int             `json:"size"`
	Members []ClusterMember `json:"members"`
}

// ClusterReport is every clone cluster found under a tree.
type ClusterReport struct {
	Root      string    `json:"root"`
	Functions int       `json:"functions"`
	Threshold float64   `json:"threshold"`
	Clusters  []Cluster `json:"clusters"`
}

// Clusters groups the functions under root into clone families: a cluster is a
// connected component of the graph where an edge joins two functions whose
// structural similarity is at or above threshold. Only clusters with more than
// one member are returned, largest first.
func Clusters(root string, threshold float64) (ClusterReport, error) {
	funcs, err := srcwalk.Walk(root)
	if err != nil {
		return ClusterReport{}, err
	}
	prints := make([]fingerprint.Print, len(funcs))
	for i := range funcs {
		prints[i] = fingerprint.New(funcs[i].Decl)
	}

	uf := newUnionFind(len(funcs))
	for i := 0; i < len(funcs); i++ {
		for j := i + 1; j < len(funcs); j++ {
			if fingerprint.Similarity(prints[i], prints[j]) >= threshold {
				uf.union(i, j)
			}
		}
	}

	groups := map[int][]int{}
	for i := range funcs {
		root := uf.find(i)
		groups[root] = append(groups[root], i)
	}

	report := ClusterReport{Root: root, Functions: len(funcs), Threshold: threshold}
	id := 0
	for _, members := range groups {
		if len(members) < 2 {
			continue
		}
		id++
		cluster := Cluster{ID: id, Size: len(members)}
		for _, idx := range members {
			cluster.Members = append(cluster.Members, ClusterMember{
				Func: funcs[idx].Name, File: funcs[idx].File, Line: funcs[idx].Line,
			})
		}
		sort.Slice(cluster.Members, func(a, b int) bool { return cluster.Members[a].Func < cluster.Members[b].Func })
		report.Clusters = append(report.Clusters, cluster)
	}
	sort.SliceStable(report.Clusters, func(i, j int) bool { return report.Clusters[i].Size > report.Clusters[j].Size })
	return report, nil
}

type unionFind struct{ parent []int }

func newUnionFind(n int) *unionFind {
	uf := &unionFind{parent: make([]int, n)}
	for i := range uf.parent {
		uf.parent[i] = i
	}
	return uf
}

func (u *unionFind) find(x int) int {
	for u.parent[x] != x {
		u.parent[x] = u.parent[u.parent[x]]
		x = u.parent[x]
	}
	return x
}

func (u *unionFind) union(a, b int) { u.parent[u.find(a)] = u.find(b) }
