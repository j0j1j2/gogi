package com.gogi.loader;

import android.app.Activity;
import android.os.Bundle;
import android.widget.LinearLayout;
import android.widget.TextView;

public class MainActivity extends Activity {
    private static native int gogiTargetRead();

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        TextView title = new TextView(this);
        title.setTextSize(24);
        title.setText("gogi loader");

        TextView status = new TextView(this);
        status.setTextSize(18);
        TextView targetValue = new TextView(this);
        targetValue.setTextSize(18);

        try {
            System.loadLibrary("target");
            System.loadLibrary("gogi");
            status.setText("libgogi.so loaded");
        } catch (Throwable t) {
            status.setText("load failed: " + t.getClass().getSimpleName() + ": " + t.getMessage());
        }

        LinearLayout layout = new LinearLayout(this);
        layout.setOrientation(LinearLayout.VERTICAL);
        int padding = 48;
        layout.setPadding(padding, padding, padding, padding);
        layout.addView(title);
        layout.addView(status);
        layout.addView(targetValue);
        setContentView(layout);
        refreshTargetValue(targetValue);
    }

    private void refreshTargetValue(TextView targetValue) {
        try {
            targetValue.setText("target value: " + gogiTargetRead());
        } catch (Throwable t) {
            targetValue.setText("target read failed: " + t.getMessage());
        }
        targetValue.postDelayed(() -> refreshTargetValue(targetValue), 500);
    }
}
