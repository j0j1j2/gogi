//go:build android

package menu

/*
#include <jni.h>
#include <stdlib.h>

static int gogi_check(JNIEnv* env) {
	if ((*env)->ExceptionCheck(env)) {
		(*env)->ExceptionDescribe(env);
		(*env)->ExceptionClear(env);
		return 0;
	}
	return 1;
}

static jobject gogi_find_activity(JNIEnv* env, int* err_code) {
	jclass at_cls = (*env)->FindClass(env, "android/app/ActivityThread");
	if (at_cls == NULL || !gogi_check(env)) { *err_code = 10; return NULL; }

	jmethodID current_mid = (*env)->GetStaticMethodID(env, at_cls, "currentActivityThread", "()Landroid/app/ActivityThread;");
	if (current_mid == NULL || !gogi_check(env)) { *err_code = 11; return NULL; }

	jobject at = (*env)->CallStaticObjectMethod(env, at_cls, current_mid);
	if (at == NULL || !gogi_check(env)) { *err_code = 12; return NULL; }

	jfieldID activities_fid = (*env)->GetFieldID(env, at_cls, "mActivities", "Landroid/util/ArrayMap;");
	if (activities_fid == NULL || !gogi_check(env)) { *err_code = 13; return NULL; }

	jobject activities = (*env)->GetObjectField(env, at, activities_fid);
	if (activities == NULL || !gogi_check(env)) { *err_code = 14; return NULL; }

	jclass map_cls = (*env)->FindClass(env, "java/util/Map");
	if (map_cls == NULL || !gogi_check(env)) { *err_code = 15; return NULL; }

	jmethodID values_mid = (*env)->GetMethodID(env, map_cls, "values", "()Ljava/util/Collection;");
	if (values_mid == NULL || !gogi_check(env)) { *err_code = 16; return NULL; }

	jobject values = (*env)->CallObjectMethod(env, activities, values_mid);
	if (values == NULL || !gogi_check(env)) { *err_code = 17; return NULL; }

	jclass coll_cls = (*env)->FindClass(env, "java/util/Collection");
	if (coll_cls == NULL || !gogi_check(env)) { *err_code = 18; return NULL; }

	jmethodID to_array_mid = (*env)->GetMethodID(env, coll_cls, "toArray", "()[Ljava/lang/Object;");
	if (to_array_mid == NULL || !gogi_check(env)) { *err_code = 19; return NULL; }

	jobjectArray records = (jobjectArray)(*env)->CallObjectMethod(env, values, to_array_mid);
	if (records == NULL || !gogi_check(env)) { *err_code = 20; return NULL; }

	jsize count = (*env)->GetArrayLength(env, records);
	if (count == 0) { *err_code = 21; return NULL; }
	for (jsize i = 0; i < count; i++) {
		jobject record = (*env)->GetObjectArrayElement(env, records, i);
		if (record == NULL) continue;

		jclass record_cls = (*env)->GetObjectClass(env, record);
		if (record_cls == NULL || !gogi_check(env)) continue;

		int paused = 0;
		jfieldID paused_fid = (*env)->GetFieldID(env, record_cls, "paused", "Z");
		if (paused_fid != NULL && gogi_check(env)) {
			paused = (*env)->GetBooleanField(env, record, paused_fid);
		} else {
			(*env)->ExceptionClear(env);
		}
		if (paused) continue;

		jfieldID activity_fid = (*env)->GetFieldID(env, record_cls, "activity", "Landroid/app/Activity;");
		if (activity_fid == NULL || !gogi_check(env)) continue;

		jobject activity = (*env)->GetObjectField(env, record, activity_fid);
		if (activity != NULL && gogi_check(env)) return activity;
	}
	*err_code = 22;
	return NULL;
}

static int gogi_attach_webview(JavaVM* vm, const char* url, const char* config_json) {
	if (vm == NULL || url == NULL || config_json == NULL) return 1;

	JNIEnv* env = NULL;
	if ((*vm)->AttachCurrentThread(vm, &env, NULL) != JNI_OK || env == NULL) return 2;

	int err_code = 0;
	jobject activity = gogi_find_activity(env, &err_code);
	if (activity == NULL) return err_code;

	jclass object_cls = (*env)->FindClass(env, "java/lang/Object");
	jclass helper_cls = NULL;
	if (object_cls != NULL && gogi_check(env)) {
		jmethodID get_class_mid = (*env)->GetMethodID(env, object_cls, "getClass", "()Ljava/lang/Class;");
		if (get_class_mid != NULL && gogi_check(env)) {
			jobject activity_class = (*env)->CallObjectMethod(env, activity, get_class_mid);
			if (activity_class != NULL && gogi_check(env)) {
				jclass class_cls = (*env)->FindClass(env, "java/lang/Class");
				if (class_cls != NULL && gogi_check(env)) {
					jmethodID get_loader_mid = (*env)->GetMethodID(env, class_cls, "getClassLoader", "()Ljava/lang/ClassLoader;");
					if (get_loader_mid != NULL && gogi_check(env)) {
						jobject loader = (*env)->CallObjectMethod(env, activity_class, get_loader_mid);
						if (loader != NULL && gogi_check(env)) {
							jclass loader_cls = (*env)->FindClass(env, "java/lang/ClassLoader");
							if (loader_cls != NULL && gogi_check(env)) {
								jmethodID load_class_mid = (*env)->GetMethodID(env, loader_cls, "loadClass", "(Ljava/lang/String;)Ljava/lang/Class;");
								jstring helper_name = (*env)->NewStringUTF(env, "com.gogi.GogiOverlay");
								if (load_class_mid != NULL && helper_name != NULL && gogi_check(env)) {
									helper_cls = (jclass)(*env)->CallObjectMethod(env, loader, load_class_mid, helper_name);
									gogi_check(env);
								}
							}
						}
					}
				}
			}
		}
	}
	if (helper_cls != NULL) {
		jmethodID attach_mid = (*env)->GetStaticMethodID(env, helper_cls, "attach", "(Landroid/app/Activity;Ljava/lang/String;Ljava/lang/String;)V");
		if (attach_mid != NULL && gogi_check(env)) {
			jstring helper_url = (*env)->NewStringUTF(env, url);
			jstring helper_config = (*env)->NewStringUTF(env, config_json);
			if (helper_url == NULL || !gogi_check(env)) return 51;
			if (helper_config == NULL || !gogi_check(env)) return 53;
			(*env)->CallStaticVoidMethod(env, helper_cls, attach_mid, activity, helper_url, helper_config);
			if (!gogi_check(env)) return 52;
			return 0;
		}
	} else {
		(*env)->ExceptionClear(env);
	}

	jclass webview_cls = (*env)->FindClass(env, "android/webkit/WebView");
	if (webview_cls == NULL || !gogi_check(env)) return 30;

	jmethodID webview_ctor = (*env)->GetMethodID(env, webview_cls, "<init>", "(Landroid/content/Context;)V");
	if (webview_ctor == NULL || !gogi_check(env)) return 31;

	jobject webview = (*env)->NewObject(env, webview_cls, webview_ctor, activity);
	if (webview == NULL || !gogi_check(env)) return 32;

	jmethodID set_bg_mid = (*env)->GetMethodID(env, webview_cls, "setBackgroundColor", "(I)V");
	if (set_bg_mid != NULL && gogi_check(env)) {
		(*env)->CallVoidMethod(env, webview, set_bg_mid, 0x00000000);
		gogi_check(env);
	}

	jmethodID get_settings_mid = (*env)->GetMethodID(env, webview_cls, "getSettings", "()Landroid/webkit/WebSettings;");
	if (get_settings_mid != NULL && gogi_check(env)) {
		jobject settings = (*env)->CallObjectMethod(env, webview, get_settings_mid);
		if (settings != NULL && gogi_check(env)) {
			jclass settings_cls = (*env)->FindClass(env, "android/webkit/WebSettings");
			if (settings_cls != NULL && gogi_check(env)) {
				jmethodID set_js_mid = (*env)->GetMethodID(env, settings_cls, "setJavaScriptEnabled", "(Z)V");
				if (set_js_mid != NULL && gogi_check(env)) {
					(*env)->CallVoidMethod(env, settings, set_js_mid, JNI_TRUE);
					gogi_check(env);
				}
			}
		}
	}

	jstring jurl = (*env)->NewStringUTF(env, url);
	if (jurl == NULL || !gogi_check(env)) return 33;
	jmethodID load_url_mid = (*env)->GetMethodID(env, webview_cls, "loadUrl", "(Ljava/lang/String;)V");
	if (load_url_mid == NULL || !gogi_check(env)) return 34;
	(*env)->CallVoidMethod(env, webview, load_url_mid, jurl);
	if (!gogi_check(env)) return 35;

	jclass params_cls = (*env)->FindClass(env, "android/widget/FrameLayout$LayoutParams");
	if (params_cls == NULL || !gogi_check(env)) return 36;

	jmethodID params_ctor = (*env)->GetMethodID(env, params_cls, "<init>", "(II)V");
	if (params_ctor == NULL || !gogi_check(env)) return 37;

	jobject params = (*env)->NewObject(env, params_cls, params_ctor, 720, 560);
	if (params == NULL || !gogi_check(env)) return 38;

	jfieldID gravity_fid = (*env)->GetFieldID(env, params_cls, "gravity", "I");
	if (gravity_fid != NULL && gogi_check(env)) {
		(*env)->SetIntField(env, params, gravity_fid, 0x35);
	}

	jclass margin_cls = (*env)->FindClass(env, "android/view/ViewGroup$MarginLayoutParams");
	if (margin_cls != NULL && gogi_check(env)) {
		jmethodID set_margins_mid = (*env)->GetMethodID(env, margin_cls, "setMargins", "(IIII)V");
		if (set_margins_mid != NULL && gogi_check(env)) {
			(*env)->CallVoidMethod(env, params, set_margins_mid, 0, 120, 24, 0);
			gogi_check(env);
		}
	}

	jclass activity_cls = (*env)->FindClass(env, "android/app/Activity");
	if (activity_cls == NULL || !gogi_check(env)) return 39;

	jmethodID add_mid = (*env)->GetMethodID(env, activity_cls, "addContentView", "(Landroid/view/View;Landroid/view/ViewGroup$LayoutParams;)V");
	if (add_mid == NULL || !gogi_check(env)) return 40;

	(*env)->CallVoidMethod(env, activity, add_mid, webview, params);
	if (!gogi_check(env)) return 41;
	return 0;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

var ErrAutoAttachUnavailable = errors.New("auto_attach_unavailable")

func AttachAuto(vm unsafe.Pointer, url string, configJSON string) error {
	cURL := C.CString(url)
	defer C.free(unsafe.Pointer(cURL))
	cConfig := C.CString(configJSON)
	defer C.free(unsafe.Pointer(cConfig))
	if code := int(C.gogi_attach_webview((*C.JavaVM)(vm), cURL, cConfig)); code != 0 {
		return fmt.Errorf("%w:%d", ErrAutoAttachUnavailable, code)
	}
	return nil
}
