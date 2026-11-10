package queue
import("context";"os";"testing")
func TestRedisListQueueRoundTrip(t *testing.T){
 addr:=os.Getenv("CATALOG_TEST_REDIS_ADDR");if addr==""{t.Skip("set dedicated disposable Redis address")}
 q,err:=New(addr);if err!=nil{t.Fatal(err)};defer q.Close();ctx:=context.Background()
 // This test requires a dedicated empty Redis instance.
 for _,id:=range []string{"11111111-1111-4111-8111-111111111111","22222222-2222-4222-8222-222222222222"}{if err=q.Enqueue(ctx,id);err!=nil{t.Fatal(err)}}
 for _,want:=range []string{"11111111-1111-4111-8111-111111111111","22222222-2222-4222-8222-222222222222"}{got,err:=q.Dequeue(ctx);if err!=nil||got!=want{t.Fatalf("FIFO got %s want %s err %v",got,want,err)}}
}
