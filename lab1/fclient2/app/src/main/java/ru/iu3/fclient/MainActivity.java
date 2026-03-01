package ru.iu3.fclient;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.Toast;

import androidx.activity.result.ActivityResultLauncher;
import androidx.activity.result.contract.ActivityResultContracts;
import androidx.appcompat.app.AppCompatActivity;

import org.apache.commons.codec.DecoderException;
import org.apache.commons.codec.binary.Hex;

import java.nio.charset.StandardCharsets;
import java.util.Arrays;

public class MainActivity extends AppCompatActivity implements TransactionEvents {

    private static final String TAG = "LIBS";

    private String pin = "";
    private ActivityResultLauncher<Intent> pinpadLauncher;

    static {
        try {
            System.loadLibrary("mbedcrypto");
            Log.e(TAG, "mbedcrypto loaded OK");
        } catch (Throwable t) {
            Log.e(TAG, "mbedcrypto load FAILED", t);
        }

        try {
            System.loadLibrary("fclient2");
            Log.e(TAG, "fclient2 loaded OK");
        } catch (Throwable t) {
            Log.e(TAG, "fclient2 load FAILED", t);
        }
    }

    // JNI-методы
    public native String stringFromJNI();
    public native boolean transaction(byte[] trd);

    // RNG
    public static native int initRng();
    public static native byte[] randomBytes(int no);

    // 3DES
    public static native byte[] encrypt(byte[] key, byte[] data);
    public static native byte[] decrypt(byte[] key, byte[] data);

    @Override
    protected void onCreate(Bundle savedInstanceState) {

        Log.e(TAG, "Calling stringFromJNI()");
        String msg = stringFromJNI();
        Log.e(TAG, "stringFromJNI returned: " + msg);

        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        Log.e(TAG, "APP START onCreate " + System.currentTimeMillis());


        //Регистрируем обработчик результата PinpadActivity
        pinpadLauncher = registerForActivityResult(
                new ActivityResultContracts.StartActivityForResult(),
                result -> {
                    if (result.getResultCode() == Activity.RESULT_OK && result.getData() != null) {
                        String received = result.getData().getStringExtra("pin");
                        pin = (received != null) ? received : "";
                    } else {
                        pin = "";
                    }

                    synchronized (MainActivity.this) {
                        MainActivity.this.notifyAll();
                    }
                }
        );

        //RNG test
        int res = initRng();
        Log.e(TAG, "initRng result = " + res);

        byte[] v = randomBytes(10);
        Log.e(TAG, "randomBytes(10) = " + Arrays.toString(v));

        //3DES test
        byte[] key = randomBytes(16); // set2key -> 16 bytes
        byte[] data = new byte[8];     // 8 байт (кратность 8)
        byte[] src = "3DES".getBytes(StandardCharsets.UTF_8);
        System.arraycopy(src, 0, data, 0, src.length);

        byte[] enc = encrypt(key, data);
        Log.e(TAG, "enc = " + Arrays.toString(enc));

        byte[] dec = decrypt(key, enc);

        int end = 0;
        while (end < dec.length && dec[end] != 0) end++;
        Log.e(TAG, "dec str = " + new String(dec, 0, end, StandardCharsets.UTF_8));
    }

    public static byte[] stringToHex(String s) {
        try {
            return Hex.decodeHex(s.toCharArray());
        } catch (DecoderException ex) {
            return null;
        }
    }

    public void onButtonClick(View v) {

        //3DES тест
        byte[] key = stringToHex("0123456789ABCDEF0123456789ABCDE0");
        byte[] data = stringToHex("0000000000000102");

        if (key == null || data == null) {
            Toast.makeText(this, "HEX decode error", Toast.LENGTH_SHORT).show();
            return;
        }

        byte[] enc = encrypt(key, data);
        if (enc == null) {
            Toast.makeText(this, "encrypt() returned null", Toast.LENGTH_SHORT).show();
            return;
        }

        byte[] dec = decrypt(key, enc);
        if (dec == null) {
            Toast.makeText(this, "decrypt() returned null", Toast.LENGTH_SHORT).show();
            return;
        }

        String s = new String(Hex.encodeHex(dec)).toUpperCase();
        Toast.makeText(this, s, Toast.LENGTH_SHORT).show();

        byte[] trd = stringToHex("9F0206000000000100");
        if (trd == null) {
            Toast.makeText(this, "TRD HEX decode error", Toast.LENGTH_SHORT).show();
            return;
        }

        transaction(trd);
    }
    @Override
    public String enterPin(int ptc, String amount) {
        pin = "";

        Intent it = new Intent(MainActivity.this, PinpadActivity.class);
        it.putExtra("ptc", ptc);
        it.putExtra("amount", amount);

        synchronized (MainActivity.this) {
            pinpadLauncher.launch(it);
            try {
                MainActivity.this.wait();
            } catch (InterruptedException ex) {
                pin = "";
            }
        }
        return pin;
    }

    @Override
    public void transactionResult(boolean result) {
        runOnUiThread(() -> {
            Toast.makeText(MainActivity.this, result ? "ok" : "failed", Toast.LENGTH_SHORT).show();
        });
    }
}