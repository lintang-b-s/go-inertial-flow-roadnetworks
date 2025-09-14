package partitioner

// import (
// 	"log"
// 	"math"
// 	"sort"

// 	"github.com/lintang-b-s/go-graph-inertial-flow/pkg/datastructure"
// 	"gonum.org/v1/gonum/mat"
// )

// gagal (gonum out-of-memory & too slow)

// graph representation learning, hamilton et al
// 
// type SpectralPartitioner struct {
// }

// func NewSpectralPartitioner(n int32) *SpectralPartitioner {
// 	return &SpectralPartitioner{}
// }

// func (sp *SpectralPartitioner) RunSpectralPartition(graph *datastructure.Graph) {

// 	// mat.NewDense(n,n,nil) out-of-memory
// 	adjacencyMat := make([][]float64, len(graph.GetNodes()))
// 	for i := range adjacencyMat {
// 		adjacencyMat[i] = make([]float64, len(adjacencyMat))
// 	}
// 	for _, edge := range graph.GraphStorage.EdgeStorage {
// 		if !edge.Directed {
// 			adjacencyMat[edge.FromNodeID][edge.ToNodeID] = edge.Weight
// 			adjacencyMat[edge.ToNodeID][edge.FromNodeID] = edge.Weight
// 		} else {
// 			adjacencyMat[edge.FromNodeID][edge.ToNodeID] = edge.Weight
// 		}
// 	}

// 	n := len(graph.GetNodes())

// 	// denseMat := mat.NewDense(n, n, nil)
// 	// for row := range adjacencyMat {
// 	// 	for col := range adjacencyMat[row] {
// 	// 		// https://pkg.go.dev/gonum.org/v1/gonum/mat#NewDense
// 	// 		denseMat.Set(row, col, adjacencyMat[row][col])
// 	// 	}
// 	// }
// 	// denseMatTranspose := denseMat.T()

// 	// // create symmetric adjacency matrix
// 	// var symMat mat.Dense
// 	// symMat.Add(denseMat, denseMatTranspose)

// 	// var scaledSymMat mat.Dense
// 	// scaledSymMat.Scale(0.5, &symMat)

// 	adjMatT := transpose(adjacencyMat)

// 	symMat := make([][]float64, n)
// 	for i := range symMat {
// 		symMat[i] = make([]float64, n)
// 	}

// 	for i := 0; i < n; i++ {
// 		for j := 0; j < n; j++ {
// 			symMat[i][j] = (adjacencyMat[i][j] + adjMatT[i][j]) * 0.5
// 		}
// 	}

// 	// build normalized symmetric laplacian matrix

// 	// diagonal matrix
// 	// D := mat.NewDense(n, n, nil)
// 	// for i := 0; i < n; i++ {
// 	// 	sum := 0.0
// 	// 	row := scaledSymMat.RawRowView(i)
// 	// 	for j := 0; j < n; j++ {
// 	// 		sum += row[j]
// 	// 	}
// 	// 	D.Set(i, i, sum)
// 	// }

// 	D := make([][]float64, n)
// 	for i := range D {
// 		D[i] = make([]float64, n)
// 	}

// 	for i := 0; i < n; i++ {
// 		sum := 0.0
// 		for j := 0; j < n; j++ {
// 			sum += symMat[i][j]
// 		}
// 		D[i][i] = sum
// 	}

// 	//  construct normalized symmetric laplacian matrix
// 	// L = I - D^(-1/2) * A * D^(-1/2)

// 	log.Printf("constructing normalized symmetric laplacian matrix")
// 	// laplacian := mat.NewDense(int(n), int(n), nil)
// 	laplacian := make([][]float64, n)
// 	for i := range laplacian {
// 		laplacian[i] = make([]float64, n)
// 	}

// 	for i := 0; i < n; i++ {
// 		for j := 0; j < n; j++ {
// 			laplacian[i][j] = -symMat[i][j]
// 		}
// 		laplacian[i][i] = D[i][i]
// 	}

// 	// for i := 0; i < n; i++ {
// 	// 	symMatRow := scaledSymMat.RawRowView(i)
// 	// 	laplacian.SetRow(i, symMatRow)

// 	// 	diagElem := D.At(i, i)
// 	// 	laplacian.Set(i, i, diagElem)
// 	// }

// 	// symLaplacian := mat.NewSymDense(int(n), nil)
// 	// for i := 0; i < n; i++ {
// 	// 	for j := 0; j < n; j++ {
// 	// 		if D[i][i] > 0 && D[i][j] > 0 {
// 	// 			val := laplacian.At(i, j) / math.Sqrt(D.At(i, i)*D.At(j, j))
// 	// 			symLaplacian.SetSym(i, j, val)
// 	// 		}
// 	// 	}
// 	// }

// 	symLaplacian := mat.NewSymDense(int(n), nil)
// 	for i := 0; i < n; i++ {
// 		for j := 0; j < n; j++ {
// 			if D[i][i] > 0 && D[i][j] > 0 {
// 				val := laplacian[i][j] / math.Sqrt(D[i][i]*D[j][j])
// 				symLaplacian.SetSym(i, j, val)
// 			}
// 		}
// 	}

// 	// eigendecomposition of symmetric normalized laplacian matrix L
// 	log.Printf("performing eigendecomposition")

// 	var eig mat.EigenSym
// 	ok := eig.Factorize(symLaplacian, true)
// 	if !ok {
// 		log.Fatal("Symmetric eigendecomposition failed")
// 	}

// 	eigenValues := eig.Values(nil)
// 	var eigenVectors mat.Dense
// 	eig.VectorsTo(&eigenVectors)

// 	// sort and get k eigenvectors with smallest eigenvalues
// 	type eigenPair struct {
// 		eigenValue  float64
// 		eigenVector mat.Vector
// 		vectorID    int
// 	}

// 	eigenPairs := make([]eigenPair, len(eigenValues))
// 	for i, val := range eigenValues {
// 		vec := eigenVectors.ColView(i)
// 		eigenPairs[i] = eigenPair{
// 			eigenValue:  val,
// 			eigenVector: vec,
// 			vectorID:    i,
// 		}
// 	}

// 	// sort by eigenvalue
// 	sort.Slice(eigenPairs, func(i, j int) bool {
// 		return eigenPairs[i].eigenValue < eigenPairs[j].eigenValue
// 	})

// 	// eigenhap heurstics to get optimal number of partitions (k)
// 	difSlice := make([]float64, len(eigenPairs)-1)
// 	for i := 0; i < len(eigenPairs)-1; i++ {
// 		dif := eigenPairs[i+1].eigenValue - eigenPairs[i].eigenValue
// 		difSlice[i] = dif
// 	}

// 	// argmax
// 	k := 1 // number of partitions
// 	maxDif := difSlice[0]
// 	for i := 1; i < len(difSlice); i++ {
// 		if difSlice[i] > maxDif {
// 			maxDif = difSlice[i]
// 			k = i
// 		}
// 	}

// 	log.Printf("performing spectral partitioning with k=%d", k)
// 	// construct matrix U that contains first k eigenvectors as columns
// 	U := mat.NewDense(int(n), k, nil)
// 	for i := 0; i < k; i++ {
// 		vec := eigenPairs[i].eigenVector
// 		columnVector := make([]float64, n)
// 		for j := 0; j < n; j++ {
// 			columnVector[j] = vec.AtVec(j)
// 		}
// 		U.SetCol(i, columnVector)
// 	}

// 	// normalize the rows of U to have unit 2-norm
// 	for i := 0; i < n; i++ {
// 		row := U.RawRowView(i)
// 		norm := 0.0
// 		for j := 0; j < k; j++ {
// 			norm += row[j] * row[j]
// 		}
// 		norm = math.Sqrt(norm)
// 		if norm > 0 {
// 			for j := 0; j < k; j++ {
// 				row[j] /= norm
// 			}
// 			U.SetRow(i, row)
// 		}
// 	}

// 	kmeans := NewKmeans(n, k, U)
// 	kmeans.fit(1e-6)

// 	partitions := make([][]int32, k)
// 	r := kmeans.eStep(kmeans.computeDistances()) // (N,K)
// 	// set partition attribute to each node
// 	for i := 0; i < n; i++ {
// 		maxIndex := 0
// 		maxVal := r.At(i, 0)
// 		for j := 1; j < k; j++ {
// 			if r.At(i, j) > maxVal {
// 				maxVal = r.At(i, j)
// 				maxIndex = j
// 			}
// 		}
// 		partitions[maxIndex] = append(partitions[maxIndex], int32(i))
// 	}
// }

// func transpose(a [][]float64) [][]float64 {
// 	n := len(a)
// 	if n == 0 {
// 		return nil
// 	}
// 	m := len(a[0])
// 	t := make([][]float64, m)
// 	for i := range t {
// 		t[i] = make([]float64, n)
// 	}
// 	for i := 0; i < n; i++ {
// 		for j := 0; j < m; j++ {
// 			t[j][i] = a[i][j]
// 		}
// 	}
// 	return t
// }
