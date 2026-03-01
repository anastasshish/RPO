#include <jni.h>
#include <string>
#include <cstring>
#include <cstdint>
#include <algorithm>
#include <android/log.h>
#include <thread>

#include <spdlog/spdlog.h>
#include <spdlog/sinks/android_sink.h>

#include <mbedtls/entropy.h>
#include <mbedtls/ctr_drbg.h>
#include <mbedtls/des.h>

#define LOG_INFO(...) __android_log_print(ANDROID_LOG_INFO, "fclient_ndk", __VA_ARGS__)
#define SLOG_INFO(...) android_logger->info(__VA_ARGS__)

// Логгер spdlog
auto android_logger = spdlog::android_logger_mt("android", "fclient_ndk");

// mbedtls RNG globals
static mbedtls_entropy_context entropy;
static mbedtls_ctr_drbg_context ctr_drbg;
static const char* personalization = "fclient-sample-app";

static JavaVM* gJvm = nullptr;

JNIEXPORT jint JNICALL JNI_OnLoad(JavaVM* pjvm, void*) {
    gJvm = pjvm;
    return JNI_VERSION_1_6;
}

static JNIEnv* getEnv(bool& detach) {
    detach = false;
    if (gJvm == nullptr) return nullptr;

    JNIEnv* env = nullptr;
    jint st = gJvm->GetEnv(reinterpret_cast<void**>(&env), JNI_VERSION_1_6);

    if (st == JNI_EDETACHED) {
        if (gJvm->AttachCurrentThread(&env, nullptr) == JNI_OK) {
            detach = true;
            return env;
        }
        return nullptr;
    }
    if (st == JNI_OK) return env;
    return nullptr;
}

static void releaseEnv(bool detach, JNIEnv*) {
    if (detach && gJvm != nullptr) {
        gJvm->DetachCurrentThread();
    }
}

extern "C" JNIEXPORT jstring JNICALL
Java_ru_iu3_fclient_MainActivity_stringFromJNI(JNIEnv* env, jobject /* this */) {
    android_logger->set_pattern("[%n] [%l] %v");

    std::string hello = "Hello from C++";
    LOG_INFO("Hello from c++ %d", 2023);
    SLOG_INFO("Hello from spdlog {}", 2023);

    return env->NewStringUTF(hello.c_str());
}

// чтобы потом randomBytes() работал
extern "C" JNIEXPORT jint JNICALL
Java_ru_iu3_fclient_MainActivity_initRng(JNIEnv* env, jclass clazz) {
    (void)env;
    (void)clazz;

    mbedtls_entropy_init(&entropy);
    mbedtls_ctr_drbg_init(&ctr_drbg);

    return mbedtls_ctr_drbg_seed(
            &ctr_drbg,
            mbedtls_entropy_func,
            &entropy,
            reinterpret_cast<const unsigned char*>(personalization),
            std::strlen(personalization)
    );
}


extern "C" JNIEXPORT jbyteArray JNICALL
Java_ru_iu3_fclient_MainActivity_randomBytes(JNIEnv* env, jclass clazz, jint no) {
    (void)clazz;

    if (no <= 0) {
        return env->NewByteArray(0);
    }

    uint8_t* buf = new uint8_t[no];
    mbedtls_ctr_drbg_random(&ctr_drbg, buf, static_cast<size_t>(no));

    jbyteArray rnd = env->NewByteArray(no);
    env->SetByteArrayRegion(rnd, 0, no, reinterpret_cast<jbyte*>(buf));

    delete[] buf;
    return rnd;
}

// шифрование
extern "C" JNIEXPORT jbyteArray JNICALL
Java_ru_iu3_fclient_MainActivity_encrypt(JNIEnv *env, jclass, jbyteArray key, jbyteArray data)
{
    jsize ksz = env->GetArrayLength(key);
    jsize dsz = env->GetArrayLength(data);
    if ((ksz != 16) || (dsz <= 0)) {
        return env->NewByteArray(0);
    }

    mbedtls_des3_context ctx;
    mbedtls_des3_init(&ctx);

    jbyte *pkey  = env->GetByteArrayElements(key, 0);
    jbyte *pdata = env->GetByteArrayElements(data, 0);


    int rst = dsz % 8;
    int pad = 8 - rst;
    int sz  = dsz + pad;

    uint8_t *buf = new uint8_t[sz];


    std::copy(reinterpret_cast<uint8_t*>(pdata),
              reinterpret_cast<uint8_t*>(pdata) + dsz,
              buf);


    for (int i = 0; i < pad; i++) {
        buf[dsz + i] = static_cast<uint8_t>(pad);
    }

    mbedtls_des3_set2key_enc(&ctx, reinterpret_cast<uint8_t*>(pkey));

    int cn = sz / 8;
    for (int i = 0; i < cn; i++) {
        mbedtls_des3_crypt_ecb(&ctx, buf + i * 8, buf + i * 8);
    }

    jbyteArray out = env->NewByteArray(sz);
    env->SetByteArrayRegion(out, 0, sz, reinterpret_cast<jbyte*>(buf));

    delete[] buf;
    env->ReleaseByteArrayElements(key, pkey, 0);
    env->ReleaseByteArrayElements(data, pdata, 0);

    mbedtls_des3_free(&ctx);
    return out;
}

extern "C" JNIEXPORT jbyteArray JNICALL
Java_ru_iu3_fclient_MainActivity_decrypt(JNIEnv *env, jclass, jbyteArray key, jbyteArray data)
{
    jsize ksz = env->GetArrayLength(key);
    jsize dsz = env->GetArrayLength(data);
    if ((ksz != 16) || (dsz <= 0) || ((dsz % 8) != 0)) {
        return env->NewByteArray(0);
    }

    mbedtls_des3_context ctx;
    mbedtls_des3_init(&ctx);

    jbyte *pkey  = env->GetByteArrayElements(key, 0);
    jbyte *pdata = env->GetByteArrayElements(data, 0);

    uint8_t *buf = new uint8_t[dsz];
    std::copy(reinterpret_cast<uint8_t*>(pdata),
              reinterpret_cast<uint8_t*>(pdata) + dsz,
              buf);

    mbedtls_des3_set2key_dec(&ctx, reinterpret_cast<uint8_t*>(pkey));

    int cn = dsz / 8;
    for (int i = 0; i < cn; i++) {
        mbedtls_des3_crypt_ecb(&ctx, buf + i * 8, buf + i * 8);
    }

    int pad = buf[dsz - 1];
    int sz = dsz - pad;
    if (pad <= 0 || pad > 8 || sz < 0) {
        delete[] buf;
        env->ReleaseByteArrayElements(key, pkey, 0);
        env->ReleaseByteArrayElements(data, pdata, 0);
        mbedtls_des3_free(&ctx);
        return env->NewByteArray(0);
    }

    jbyteArray out = env->NewByteArray(sz);
    env->SetByteArrayRegion(out, 0, sz, reinterpret_cast<jbyte*>(buf));

    delete[] buf;
    env->ReleaseByteArrayElements(key, pkey, 0);
    env->ReleaseByteArrayElements(data, pdata, 0);

    mbedtls_des3_free(&ctx);
    return out;
}

//обработка пина
static jboolean transaction_impl(JNIEnv* env, jobject thiz, jbyteArray trd) {
    if (!env || !thiz || trd == nullptr) return JNI_FALSE;

    jclass cls = env->GetObjectClass(thiz);
    if (!cls) return JNI_FALSE;

    jmethodID midEnterPin = env->GetMethodID(
            cls, "enterPin", "(ILjava/lang/String;)Ljava/lang/String;");
    if (!midEnterPin) return JNI_FALSE;

    uint8_t* p = reinterpret_cast<uint8_t*>(env->GetByteArrayElements(trd, nullptr));
    jsize sz = env->GetArrayLength(trd);

    if ((sz != 9) || (p[0] != 0x9F) || (p[1] != 0x02) || (p[2] != 0x06)) {
        env->ReleaseByteArrayElements(trd, reinterpret_cast<jbyte*>(p), 0);
        return JNI_FALSE;
    }

    char buf[13];
    for (int i = 0; i < 6; i++) {
        uint8_t n = *(p + 3 + i);
        buf[i * 2]     = ((n & 0xF0) >> 4) + '0';
        buf[i * 2 + 1] = (n & 0x0F) + '0';
    }
    buf[12] = 0x00;

    jstring jamount = env->NewStringUTF(buf);

    int ptc = 3;
    while (ptc > 0) {
        jstring jpin = (jstring) env->CallObjectMethod(thiz, midEnterPin, ptc, jamount);

        if (env->ExceptionCheck()) {
            env->ExceptionClear();
            env->DeleteLocalRef(jamount);
            env->ReleaseByteArrayElements(trd, reinterpret_cast<jbyte*>(p), 0);
            return JNI_FALSE;
        }

        const char* utf = (jpin != nullptr) ? env->GetStringUTFChars(jpin, nullptr) : nullptr;
        bool ok = (utf != nullptr) && (std::strcmp(utf, "1234") == 0);

        if (jpin != nullptr && utf != nullptr) env->ReleaseStringUTFChars(jpin, utf);
        if (jpin != nullptr) env->DeleteLocalRef(jpin);

        if (ok) break;
        ptc--;
    }

    env->DeleteLocalRef(jamount);
    env->ReleaseByteArrayElements(trd, reinterpret_cast<jbyte*>(p), 0);

    return (ptc > 0) ? JNI_TRUE : JNI_FALSE;
}


extern "C"
JNIEXPORT jboolean JNICALL
Java_ru_iu3_fclient_MainActivity_transaction(JNIEnv *xenv, jobject xthiz, jbyteArray xtrd) {

    if (gJvm == nullptr || xenv == nullptr || xthiz == nullptr || xtrd == nullptr) return JNI_FALSE;

    jobject thiz = xenv->NewGlobalRef(xthiz);
    jbyteArray trd = (jbyteArray) xenv->NewGlobalRef(xtrd);

    std::thread t([thiz, trd]() {
        bool detach = false;
        JNIEnv* env = getEnv(detach);
        if (!env) return;

        jboolean ok = transaction_impl(env, thiz, trd);

        jclass cls = env->GetObjectClass(thiz);
        if (cls) {
            jmethodID midRes = env->GetMethodID(cls, "transactionResult", "(Z)V");
            if (midRes) {
                env->CallVoidMethod(thiz, midRes, ok);
                if (env->ExceptionCheck()) env->ExceptionClear();
            }
        }

        env->DeleteGlobalRef(thiz);
        env->DeleteGlobalRef(trd);

        releaseEnv(detach, env);
    });

    t.detach();
    return JNI_TRUE;
}