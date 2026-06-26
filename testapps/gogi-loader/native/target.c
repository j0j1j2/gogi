#include <stdint.h>
#include <jni.h>

__attribute__((visibility("default")))
int32_t gogi_target_value = 7;

__attribute__((visibility("default")))
int32_t gogi_target_read(void) {
    return gogi_target_value;
}

JNIEXPORT jint JNICALL
Java_com_gogi_loader_MainActivity_gogiTargetRead(JNIEnv *env, jclass clazz) {
    (void)env;
    (void)clazz;
    return gogi_target_read();
}
