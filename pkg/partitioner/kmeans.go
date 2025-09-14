package partitioner

// import (
// 	"math"

// 	"gonum.org/v1/gonum/mat"
// )

// type Kmeans struct {
// 	N              int
// 	M              int
// 	K              int
// 	X              *mat.Dense // (N,M),
// 	i              int
// 	Prototypes     *mat.Dense // (K,M)
// 	distortionHist []float64
// }

// func NewKmeans(n, k int, x *mat.Dense) *Kmeans {
// 	// init random center with dimension (K,M)
// 	m := x.RawMatrix().Cols
// 	prototypes := mat.NewDense(k, m, nil)

// 	km := &Kmeans{
// 		N:              n,
// 		M:              m,
// 		K:              k,
// 		X:              x,
// 		i:              0,
// 		Prototypes:     prototypes,
// 		distortionHist: []float64{},
// 	}
// 	km.distortionHist = append(km.distortionHist, km.computeDistortion())

// 	return km
// }

// func (km *Kmeans) computeDistances() [][]*mat.VecDense {

// 	distanceMat := make([][]*mat.VecDense, km.N)
// 	for i := 0; i < km.N; i++ {
// 		xi := mat.Row(nil, i, km.X)
// 		vXi := mat.NewVecDense(km.M, xi)

// 		xiClustDist := make([]*mat.VecDense, km.K)

// 		for k := 0; k < km.K; k++ {
// 			protoK := mat.Row(nil, k, km.Prototypes)
// 			vProtoK := mat.NewVecDense(km.M, protoK)

// 			diff := mat.NewVecDense(km.M, nil)
// 			diff.SubVec(vXi, vProtoK)
// 			diff.MulElemVec(diff, diff)

// 			xiClustDist[k] = diff
// 		}

// 		distanceMat[i] = xiClustDist
// 	}

// 	return distanceMat // (N,K,M)
// }

// func (km *Kmeans) computeDistortion() float64 {
// 	dist := km.computeDistances() // (N,K,M)
// 	r := km.eStep(dist)           // (N,K)

// 	distSummed := mat.NewDense(km.N, km.K, nil) // (N,K)
// 	// sum dist only over last axis
// 	for i := 0; i < km.N; i++ {
// 		for k := 0; k < km.K; k++ {
// 			sum := 0.0
// 			for j := 0; j < km.M; j++ {
// 				sum += dist[i][k].AtVec(j)
// 			}
// 			distSummed.Set(i, k, sum)
// 		}
// 	}

// 	distSummed.MulElem(distSummed, r) // (N,K)

// 	J := 0.0
// 	for i := 0; i < km.N; i++ {
// 		row := distSummed.RawRowView(i)
// 		for k := 0; k < km.K; k++ {
// 			J += row[k]
// 		}
// 	}
// 	return J
// }

// func (km *Kmeans) eStep(dist [][]*mat.VecDense) *mat.Dense {

// 	R := mat.NewDense(km.N, km.K, nil)

// 	for i := 0; i < km.N; i++ {
// 		minIndex := 0
// 		minVal := 1e9
// 		for k := 0; k < km.K; k++ {
// 			sum := 0.0
// 			for j := 0; j < km.M; j++ {
// 				sum += dist[i][k].AtVec(j)
// 			}
// 			if sum < minVal {
// 				minVal = sum
// 				minIndex = k
// 			}
// 		}
// 		R.Set(i, minIndex, 1)
// 	}
// 	return R // (N,K) one-hot vector of cluster assignment
// }

// func (km *Kmeans) mStep() {
// 	r := km.eStep(km.computeDistances()) // (N,K)

// 	nk := make([]float64, km.K)
// 	for k := 0; k < km.K; k++ {
// 		sum := 0.0
// 		for i := 0; i < km.N; i++ {
// 			sum += r.At(i, k)
// 		}
// 		nk[k] = sum
// 	}

// 	// reset prototypes
// 	km.Prototypes = mat.NewDense(km.K, km.M, nil)

// 	for i := 0; i < km.N; i++ {
// 		for k := 0; k < km.K; k++ {
// 			if r.At(i, k) == 1 {
// 				xi := mat.Row(nil, i, km.X)
// 				for j := 0; j < km.M; j++ {
// 					km.Prototypes.Set(k, j, km.Prototypes.At(k, j)+xi[j]/nk[k])
// 				}
// 			}
// 		}
// 	}
// }

// func (km *Kmeans) updatePrototypes() {
// 	km.mStep()
// 	Ji := km.computeDistortion()
// 	km.distortionHist = append(km.distortionHist, Ji)
// }

// func (km *Kmeans) fit(threshold float64) {
// 	diff := 1e9
// 	for diff > threshold {
// 		km.updatePrototypes()
// 		jlast, jcurrent := km.distortionHist[len(km.distortionHist)-2], km.distortionHist[len(km.distortionHist)-1]
// 		diff = math.Abs(jlast/jcurrent - 1)
// 	}
// }
