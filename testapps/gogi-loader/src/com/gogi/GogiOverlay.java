package com.gogi;

import android.app.Activity;
import android.content.Context;
import android.graphics.Color;
import android.graphics.drawable.GradientDrawable;
import android.os.Handler;
import android.os.Looper;
import android.view.Gravity;
import android.view.MotionEvent;
import android.view.View;
import android.view.ViewConfiguration;
import android.webkit.WebView;
import android.widget.FrameLayout;
import android.widget.LinearLayout;
import android.widget.TextView;

import org.json.JSONObject;

public final class GogiOverlay {
    private static View overlay;

    private static final int DEFAULT_HANDLE_SIZE = 56;
    private static final int DEFAULT_MENU_WIDTH = 320;
    private static final int DEFAULT_MENU_HEIGHT = 420;
    private static final String HANDLE_ICON = "≡";
    private static final int HANDLE_BACKGROUND = Color.rgb(34, 34, 34);
    private static final int HANDLE_FOREGROUND = Color.WHITE;
    private static final boolean INITIAL_COLLAPSED = false;

    private GogiOverlay() {}

    public static void attach(Activity activity, String url, String configJson) {
        if (activity == null || url == null) {
            return;
        }
        OverlayConfig config = OverlayConfig.parse(activity, configJson);
        activity.runOnUiThread(() -> {
            if (overlay != null) {
                return;
            }
            overlay = createPanel(activity, url, config);
            activity.addContentView(overlay, overlay.getLayoutParams());
        });
    }

    private static View createPanel(Activity activity, String url, OverlayConfig config) {
        FrameLayout panel = new FrameLayout(activity);
        LinearLayout content = new LinearLayout(activity);
        content.setOrientation(LinearLayout.VERTICAL);
        content.setBackgroundColor(Color.TRANSPARENT);

        TextView handle = new TextView(activity);
        handle.setText(HANDLE_ICON);
        handle.setTextColor(HANDLE_FOREGROUND);
        handle.setTextSize(22);
        handle.setGravity(Gravity.CENTER);
        handle.setBackground(circle(HANDLE_BACKGROUND, config.handleSize));

        WebView webView = new WebView(activity);
        webView.setBackgroundColor(Color.TRANSPARENT);
        webView.getSettings().setJavaScriptEnabled(true);
        webView.loadUrl(url);

        new Handler(Looper.getMainLooper()).postDelayed(() -> webView.loadUrl(url), 500);

        LinearLayout.LayoutParams handleParams = new LinearLayout.LayoutParams(config.handleSize, config.handleSize);
        handleParams.leftMargin = 0;
        handleParams.bottomMargin = 8;
        content.addView(handle, handleParams);
        content.addView(webView, new LinearLayout.LayoutParams(config.width, config.height));
        panel.addView(content);

        FrameLayout.LayoutParams params = new FrameLayout.LayoutParams(config.width, config.handleSize + 8 + config.height);
        params.gravity = Gravity.TOP | Gravity.LEFT;
        params.leftMargin = dp(activity, 24);
        params.topMargin = dp(activity, 96);
        panel.setLayoutParams(params);
        handle.setOnTouchListener(new OverlayTouchListener(panel, webView, config));
        setCollapsed(panel, webView, config, INITIAL_COLLAPSED);
        return panel;
    }

    private static GradientDrawable circle(int color, int size) {
        GradientDrawable drawable = new GradientDrawable();
        drawable.setShape(GradientDrawable.RECTANGLE);
        drawable.setColor(color);
        drawable.setCornerRadius(size / 2f);
        return drawable;
    }

    private static void setCollapsed(View panel, View webView, OverlayConfig config, boolean collapsed) {
        webView.setVisibility(collapsed ? View.GONE : View.VISIBLE);
        panel.setTag(Boolean.valueOf(collapsed));
        FrameLayout.LayoutParams params = (FrameLayout.LayoutParams) panel.getLayoutParams();
        params.width = collapsed ? config.handleSize : config.width;
        params.height = collapsed ? config.handleSize : config.handleSize + 8 + config.height;
        panel.setLayoutParams(params);
    }

    private static boolean isCollapsed(View panel) {
        Object tag = panel.getTag();
        return tag instanceof Boolean && (Boolean) tag;
    }

    private static final class OverlayTouchListener implements View.OnTouchListener {
        private final View target;
        private final View body;
        private final OverlayConfig config;
        private int startLeft;
        private int startTop;
        private float downRawX;
        private float downRawY;
        private boolean dragging;
        private final int touchSlop;

        OverlayTouchListener(View target, View body, OverlayConfig config) {
            this.target = target;
            this.body = body;
            this.config = config;
            this.touchSlop = ViewConfiguration.get(target.getContext()).getScaledTouchSlop();
        }

        @Override
        public boolean onTouch(View view, MotionEvent event) {
            FrameLayout.LayoutParams params = (FrameLayout.LayoutParams) target.getLayoutParams();
            switch (event.getActionMasked()) {
                case MotionEvent.ACTION_DOWN:
                    startLeft = params.leftMargin;
                    startTop = params.topMargin;
                    downRawX = event.getRawX();
                    downRawY = event.getRawY();
                    dragging = false;
                    return true;
                case MotionEvent.ACTION_MOVE:
                    if (!config.draggable) {
                        return true;
                    }
                    int dx = Math.round(event.getRawX() - downRawX);
                    int dy = Math.round(event.getRawY() - downRawY);
                    if (!dragging && Math.hypot(dx, dy) < touchSlop) {
                        return true;
                    }
                    dragging = true;
                    params.leftMargin = startLeft + dx;
                    params.topMargin = startTop + dy;
                    clamp(params);
                    target.setLayoutParams(params);
                    return true;
                case MotionEvent.ACTION_UP:
                case MotionEvent.ACTION_CANCEL:
                    if (!dragging) {
                        setCollapsed(target, body, config, !isCollapsed(target));
                    } else {
                        view.performClick();
                    }
                    return true;
                default:
                    return false;
            }
        }

        private void clamp(FrameLayout.LayoutParams params) {
            View parent = (View) target.getParent();
            if (parent == null || parent.getWidth() == 0 || parent.getHeight() == 0) {
                return;
            }
            int maxLeft = Math.max(0, parent.getWidth() - config.handleSize);
            int maxTop = Math.max(0, parent.getHeight() - config.handleSize);
            params.leftMargin = Math.max(0, Math.min(params.leftMargin, maxLeft));
            params.topMargin = Math.max(0, Math.min(params.topMargin, maxTop));
        }
    }

    private static final class OverlayConfig {
        final int width;
        final int height;
        final int handleSize;
        final boolean draggable;

        OverlayConfig(int width, int height, int handleSize, boolean draggable) {
            this.width = width;
            this.height = height;
            this.handleSize = handleSize;
            this.draggable = draggable;
        }

        static OverlayConfig parse(Context context, String json) {
            try {
                JSONObject object = new JSONObject(json == null ? "{}" : json);
                return new OverlayConfig(
                    GogiOverlay.dp(context, object.optInt("width", DEFAULT_MENU_WIDTH)),
                    GogiOverlay.dp(context, object.optInt("height", DEFAULT_MENU_HEIGHT)),
                    GogiOverlay.dp(context, object.optInt("collapsed_size", DEFAULT_HANDLE_SIZE)),
                    object.optBoolean("draggable", true)
                );
            } catch (Exception ignored) {
                return new OverlayConfig(
                    GogiOverlay.dp(context, DEFAULT_MENU_WIDTH),
                    GogiOverlay.dp(context, DEFAULT_MENU_HEIGHT),
                    GogiOverlay.dp(context, DEFAULT_HANDLE_SIZE),
                    true
                );
            }
        }

    }

    private static int dp(Context context, int value) {
        float density = context.getResources().getDisplayMetrics().density;
        return Math.max(1, Math.round(value * density));
    }
}
