//go:build cgo
// +build cgo

package native

/*
#include <string.h>
#include <stdlib.h>
#include <stdint.h>
typedef struct kclvm_service kclvm_service;
kclvm_service * kclvm_service_new(void *f,uint64_t plugin_agent){
	kclvm_service * (*new_service)(uint64_t);
	new_service = (kclvm_service * (*)(uint64_t))f;
	return new_service(plugin_agent);
}
void kclvm_service_delete(void *f,kclvm_service * c){
	void (*delete_service)(kclvm_service *);
	delete_service = (void (*)(kclvm_service *))f;
	return delete_service(c);
}
void kclvm_service_free_string(void *f,const char * res) {
	void (*free_string)(const char *);
	free_string = (void (*)(const char *))f;
	return free_string(res);
}
const char* kclvm_service_call_with_length(void *f,kclvm_service* c,const char * method,const char * args,size_t args_len,size_t * result_len){
	const char* (*service_call_with_length)(kclvm_service*,const char *,const char *,size_t,size_t *);
	service_call_with_length = (const char* (*)(kclvm_service*,const char *,const char *,size_t,size_t *))f;
	return service_call_with_length(c,method,args,args_len,result_len);
}
*/
import "C"

// NewKclvmService takes a pluginAgent and returns a pointer of capi kclvm_service.
// pluginAgent is the address of C function pointer with type :const char * (*)(const char *method,const char *args,const char *kwargs),
// in which "method" is the fully qualified name of plugin method,
// and "args" ,"kwargs"  and return value of pluginAgent are JSON string
func NewKclvmService(pluginAgent C.uint64_t) *C.kclvm_service {
	const fnName = "kclvm_service_new"

	newServ, err := lib.GetSymbolPointer(fnName)

	if err != nil {
		panic(err)
	}

	return C.kclvm_service_new(newServ, pluginAgent)
}

// NewKclvmService releases the memory of kclvm_service
func DeleteKclvmService(serv *C.kclvm_service) {
	const fnName = "kclvm_service_delete"

	deleteServ, err := lib.GetSymbolPointer(fnName)

	if err != nil {
		panic(err)
	}

	C.kclvm_service_delete(deleteServ, serv)
}

// KclvmServiceFreeString releases the memory of  the return value of KclvmServiceCall
func KclvmServiceFreeString(str *C.char) {
	const fnName = "kclvm_service_free_string"

	freeString, err := lib.GetSymbolPointer(fnName)

	if err != nil {
		panic(err)
	}

	C.kclvm_service_free_string(freeString, str)
}

// KclvmServiceCall calls kclvm service by c api
// args should be serialized as protobuf byte stream
func KclvmServiceCall(serv *C.kclvm_service, method *C.char, args *C.char, args_len C.size_t) (*C.char, C.size_t) {
	const fnName = "kclvm_service_call_with_length"

	serviceCall, err := lib.GetSymbolPointer(fnName)

	if err != nil {
		panic(err)
	}

	var size C.size_t = C.SIZE_MAX
	buf := C.kclvm_service_call_with_length(serviceCall, serv, method, args, args_len, &size)
	return buf, size
}
