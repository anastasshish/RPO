package ru.iu3.fclient;

import android.content.Intent;
import android.os.Bundle;
import android.view.View;
import android.widget.Button;
import android.widget.TextView;

import androidx.appcompat.app.AppCompatActivity;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.text.DecimalFormat;
import androidx.activity.OnBackPressedCallback;

public class PinpadActivity extends AppCompatActivity {

    private static final int MAX_KEYS = 10;
    private static final int PIN_LEN = 4;

    private TextView tvPin;
    private TextView tvAmount;
    private TextView tvPtc;

    private final List<Button> digitButtons = new ArrayList<>();
    private String pin = "";



    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_pinpad);

        getOnBackPressedDispatcher().addCallback(this, new OnBackPressedCallback(true) {
            @Override
            public void handleOnBackPressed() {
                returnResult("");
            }
        });

        tvPin = findViewById(R.id.txtPin);

        tvAmount = findViewById(R.id.txtAmount);
        tvPtc = findViewById(R.id.txtPtc);

        TextView ta = findViewById(R.id.txtAmount);

        String amt = String.valueOf(getIntent().getStringExtra("amount"));
        if (amt == null || "null".equals(amt) || amt.isEmpty()) amt = "0";

        Long f;
        try {
            f = Long.valueOf(amt);
        } catch (NumberFormatException e) {
            f = 0L;
        }

        DecimalFormat df = new DecimalFormat("#,###,###,##0.00");
        String s = df.format(f);
        ta.setText("Сумма: " + s);

        TextView tp = findViewById(R.id.txtPtc);
        int pts = getIntent().getIntExtra("ptc", 0);
        if (pts == 2)
            tp.setText("Осталось две попытки");
        else if (pts == 1)
            tp.setText("Осталась одна попытка");

        digitButtons.add(findViewById(R.id.btnKey0));
        digitButtons.add(findViewById(R.id.btnKey1));
        digitButtons.add(findViewById(R.id.btnKey2));
        digitButtons.add(findViewById(R.id.btnKey3));
        digitButtons.add(findViewById(R.id.btnKey4));
        digitButtons.add(findViewById(R.id.btnKey5));
        digitButtons.add(findViewById(R.id.btnKey6));
        digitButtons.add(findViewById(R.id.btnKey7));
        digitButtons.add(findViewById(R.id.btnKey8));
        digitButtons.add(findViewById(R.id.btnKey9));

        shuffleKeys();
        updateMaskedPin();
    }

    private void shuffleKeys() {
        byte[] rnd = MainActivity.randomBytes(MAX_KEYS);
        if (rnd == null || rnd.length < MAX_KEYS) return;

        List<Integer> digits = new ArrayList<>();
        for (int i = 0; i < 10; i++) digits.add(i);

        for (int i = digits.size() - 1; i > 0; i--) {
            int j = (rnd[i] & 0xFF) % (i + 1);
            Collections.swap(digits, i, j);
        }

        for (int i = 0; i < digitButtons.size(); i++) {
            digitButtons.get(i).setText(String.valueOf(digits.get(i)));
        }
    }

    private void updateMaskedPin() {
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < pin.length(); i++) sb.append('*');
        tvPin.setText(sb.toString());
    }

    public void keyClick(View v) {

        Button b = (Button) v;
        String t = b.getText().toString();


        if ("C".equals(t)) {
            pin = "";
            updateMaskedPin();
            return;
        }

        if ("OK".equals(t)) {
            returnResult(pin);
            return;
        }

        if (pin.length() < PIN_LEN) {
            pin += t;
            updateMaskedPin();
        }
    }

    private void returnResult(String value) {
        Intent it = new Intent();
        it.putExtra("pin", value != null ? value : "");
        setResult(RESULT_OK, it);
        finish();
    }
}