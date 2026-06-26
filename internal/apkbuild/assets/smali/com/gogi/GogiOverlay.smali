.class public final Lcom/gogi/GogiOverlay;
.super Ljava/lang/Object;
.source "GogiOverlay.java"


# annotations
.annotation system Ldalvik/annotation/MemberClasses;
    value = {
        Lcom/gogi/GogiOverlay$OverlayConfig;,
        Lcom/gogi/GogiOverlay$OverlayTouchListener;
    }
.end annotation


# static fields
.field private static final DEFAULT_HANDLE_SIZE:I = 0x38

.field private static final DEFAULT_MENU_HEIGHT:I = 0x1a4

.field private static final DEFAULT_MENU_WIDTH:I = 0x140

.field private static final HANDLE_BACKGROUND:I

.field private static final HANDLE_FOREGROUND:I = -0x1

.field private static final HANDLE_ICON:Ljava/lang/String; = "\u2261"

.field private static final INITIAL_COLLAPSED:Z = false

.field private static overlay:Landroid/view/View;


# direct methods
.method static bridge synthetic -$$Nest$smdp(Landroid/content/Context;I)I
    .locals 0

    invoke-static {p0, p1}, Lcom/gogi/GogiOverlay;->dp(Landroid/content/Context;I)I

    move-result p0

    return p0
.end method

.method static bridge synthetic -$$Nest$smisCollapsed(Landroid/view/View;)Z
    .locals 0

    invoke-static {p0}, Lcom/gogi/GogiOverlay;->isCollapsed(Landroid/view/View;)Z

    move-result p0

    return p0
.end method

.method static bridge synthetic -$$Nest$smsetCollapsed(Landroid/view/View;Landroid/view/View;Lcom/gogi/GogiOverlay$OverlayConfig;Z)V
    .locals 0

    invoke-static {p0, p1, p2, p3}, Lcom/gogi/GogiOverlay;->setCollapsed(Landroid/view/View;Landroid/view/View;Lcom/gogi/GogiOverlay$OverlayConfig;Z)V

    return-void
.end method

.method static constructor <clinit>()V
    .locals 1

    .line 27
    const/16 v0, 0x22

    invoke-static {v0, v0, v0}, Landroid/graphics/Color;->rgb(III)I

    move-result v0

    sput v0, Lcom/gogi/GogiOverlay;->HANDLE_BACKGROUND:I

    return-void
.end method

.method private constructor <init>()V
    .locals 0

    .line 31
    invoke-direct {p0}, Ljava/lang/Object;-><init>()V

    return-void
.end method

.method public static attach(Landroid/app/Activity;Ljava/lang/String;Ljava/lang/String;)V
    .locals 1

    .line 34
    if-eqz p0, :cond_1

    if-nez p1, :cond_0

    goto :goto_0

    .line 37
    :cond_0
    invoke-static {p0, p2}, Lcom/gogi/GogiOverlay$OverlayConfig;->parse(Landroid/content/Context;Ljava/lang/String;)Lcom/gogi/GogiOverlay$OverlayConfig;

    move-result-object p2

    .line 38
    new-instance v0, Lcom/gogi/GogiOverlay$0;

    invoke-direct {v0, p0, p1, p2}, Lcom/gogi/GogiOverlay$0;-><init>(Landroid/app/Activity;Ljava/lang/String;Lcom/gogi/GogiOverlay$OverlayConfig;)V

    invoke-virtual {p0, v0}, Landroid/app/Activity;->runOnUiThread(Ljava/lang/Runnable;)V

    .line 45
    return-void

    .line 35
    :cond_1
    :goto_0
    return-void
.end method

.method private static circle(II)Landroid/graphics/drawable/GradientDrawable;
    .locals 2

    .line 85
    new-instance v0, Landroid/graphics/drawable/GradientDrawable;

    invoke-direct {v0}, Landroid/graphics/drawable/GradientDrawable;-><init>()V

    .line 86
    const/4 v1, 0x0

    invoke-virtual {v0, v1}, Landroid/graphics/drawable/GradientDrawable;->setShape(I)V

    .line 87
    invoke-virtual {v0, p0}, Landroid/graphics/drawable/GradientDrawable;->setColor(I)V

    .line 88
    int-to-float p0, p1

    const/high16 p1, 0x40000000    # 2.0f

    div-float/2addr p0, p1

    invoke-virtual {v0, p0}, Landroid/graphics/drawable/GradientDrawable;->setCornerRadius(F)V

    .line 89
    return-object v0
.end method

.method private static createPanel(Landroid/app/Activity;Ljava/lang/String;Lcom/gogi/GogiOverlay$OverlayConfig;)Landroid/view/View;
    .locals 9

    .line 48
    new-instance v0, Landroid/widget/FrameLayout;

    invoke-direct {v0, p0}, Landroid/widget/FrameLayout;-><init>(Landroid/content/Context;)V

    .line 49
    new-instance v1, Landroid/widget/LinearLayout;

    invoke-direct {v1, p0}, Landroid/widget/LinearLayout;-><init>(Landroid/content/Context;)V

    .line 50
    const/4 v2, 0x1

    invoke-virtual {v1, v2}, Landroid/widget/LinearLayout;->setOrientation(I)V

    .line 51
    const/4 v3, 0x0

    invoke-virtual {v1, v3}, Landroid/widget/LinearLayout;->setBackgroundColor(I)V

    .line 53
    new-instance v4, Landroid/widget/TextView;

    invoke-direct {v4, p0}, Landroid/widget/TextView;-><init>(Landroid/content/Context;)V

    .line 54
    const-string v5, "\u2261"

    invoke-virtual {v4, v5}, Landroid/widget/TextView;->setText(Ljava/lang/CharSequence;)V

    .line 55
    const/4 v5, -0x1

    invoke-virtual {v4, v5}, Landroid/widget/TextView;->setTextColor(I)V

    .line 56
    const/high16 v5, 0x41b00000    # 22.0f

    invoke-virtual {v4, v5}, Landroid/widget/TextView;->setTextSize(F)V

    .line 57
    const/16 v5, 0x11

    invoke-virtual {v4, v5}, Landroid/widget/TextView;->setGravity(I)V

    .line 58
    sget v5, Lcom/gogi/GogiOverlay;->HANDLE_BACKGROUND:I

    iget v6, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    invoke-static {v5, v6}, Lcom/gogi/GogiOverlay;->circle(II)Landroid/graphics/drawable/GradientDrawable;

    move-result-object v5

    invoke-virtual {v4, v5}, Landroid/widget/TextView;->setBackground(Landroid/graphics/drawable/Drawable;)V

    .line 60
    new-instance v5, Landroid/webkit/WebView;

    invoke-direct {v5, p0}, Landroid/webkit/WebView;-><init>(Landroid/content/Context;)V

    .line 61
    invoke-virtual {v5, v3}, Landroid/webkit/WebView;->setBackgroundColor(I)V

    .line 62
    invoke-virtual {v5}, Landroid/webkit/WebView;->getSettings()Landroid/webkit/WebSettings;

    move-result-object v6

    invoke-virtual {v6, v2}, Landroid/webkit/WebSettings;->setJavaScriptEnabled(Z)V

    .line 63
    invoke-virtual {v5, p1}, Landroid/webkit/WebView;->loadUrl(Ljava/lang/String;)V

    .line 65
    new-instance v2, Landroid/os/Handler;

    invoke-static {}, Landroid/os/Looper;->getMainLooper()Landroid/os/Looper;

    move-result-object v6

    invoke-direct {v2, v6}, Landroid/os/Handler;-><init>(Landroid/os/Looper;)V

    new-instance v6, Lcom/gogi/GogiOverlay$1;

    invoke-direct {v6, v5, p1}, Lcom/gogi/GogiOverlay$1;-><init>(Landroid/webkit/WebView;Ljava/lang/String;)V

    const-wide/16 v7, 0x1f4

    invoke-virtual {v2, v6, v7, v8}, Landroid/os/Handler;->postDelayed(Ljava/lang/Runnable;J)Z

    .line 67
    new-instance p1, Landroid/widget/LinearLayout$LayoutParams;

    iget v2, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    iget v6, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    invoke-direct {p1, v2, v6}, Landroid/widget/LinearLayout$LayoutParams;-><init>(II)V

    .line 68
    iput v3, p1, Landroid/widget/LinearLayout$LayoutParams;->leftMargin:I

    .line 69
    const/16 v2, 0x8

    iput v2, p1, Landroid/widget/LinearLayout$LayoutParams;->bottomMargin:I

    .line 70
    invoke-virtual {v1, v4, p1}, Landroid/widget/LinearLayout;->addView(Landroid/view/View;Landroid/view/ViewGroup$LayoutParams;)V

    .line 71
    new-instance p1, Landroid/widget/LinearLayout$LayoutParams;

    iget v6, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->width:I

    iget v7, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->height:I

    invoke-direct {p1, v6, v7}, Landroid/widget/LinearLayout$LayoutParams;-><init>(II)V

    invoke-virtual {v1, v5, p1}, Landroid/widget/LinearLayout;->addView(Landroid/view/View;Landroid/view/ViewGroup$LayoutParams;)V

    .line 72
    invoke-virtual {v0, v1}, Landroid/widget/FrameLayout;->addView(Landroid/view/View;)V

    .line 74
    new-instance p1, Landroid/widget/FrameLayout$LayoutParams;

    iget v1, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->width:I

    iget v6, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    add-int/2addr v6, v2

    iget v2, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->height:I

    add-int/2addr v6, v2

    invoke-direct {p1, v1, v6}, Landroid/widget/FrameLayout$LayoutParams;-><init>(II)V

    .line 75
    const/16 v1, 0x33

    iput v1, p1, Landroid/widget/FrameLayout$LayoutParams;->gravity:I

    .line 76
    const/16 v1, 0x18

    invoke-static {p0, v1}, Lcom/gogi/GogiOverlay;->dp(Landroid/content/Context;I)I

    move-result v1

    iput v1, p1, Landroid/widget/FrameLayout$LayoutParams;->leftMargin:I

    .line 77
    const/16 v1, 0x60

    invoke-static {p0, v1}, Lcom/gogi/GogiOverlay;->dp(Landroid/content/Context;I)I

    move-result p0

    iput p0, p1, Landroid/widget/FrameLayout$LayoutParams;->topMargin:I

    .line 78
    invoke-virtual {v0, p1}, Landroid/widget/FrameLayout;->setLayoutParams(Landroid/view/ViewGroup$LayoutParams;)V

    .line 79
    new-instance p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;

    invoke-direct {p0, v0, v5, p2}, Lcom/gogi/GogiOverlay$OverlayTouchListener;-><init>(Landroid/view/View;Landroid/view/View;Lcom/gogi/GogiOverlay$OverlayConfig;)V

    invoke-virtual {v4, p0}, Landroid/widget/TextView;->setOnTouchListener(Landroid/view/View$OnTouchListener;)V

    .line 80
    invoke-static {v0, v5, p2, v3}, Lcom/gogi/GogiOverlay;->setCollapsed(Landroid/view/View;Landroid/view/View;Lcom/gogi/GogiOverlay$OverlayConfig;Z)V

    .line 81
    return-object v0
.end method

.method private static dp(Landroid/content/Context;I)I
    .locals 0

    .line 210
    invoke-virtual {p0}, Landroid/content/Context;->getResources()Landroid/content/res/Resources;

    move-result-object p0

    invoke-virtual {p0}, Landroid/content/res/Resources;->getDisplayMetrics()Landroid/util/DisplayMetrics;

    move-result-object p0

    iget p0, p0, Landroid/util/DisplayMetrics;->density:F

    .line 211
    int-to-float p1, p1

    mul-float/2addr p1, p0

    invoke-static {p1}, Ljava/lang/Math;->round(F)I

    move-result p0

    const/4 p1, 0x1

    invoke-static {p1, p0}, Ljava/lang/Math;->max(II)I

    move-result p0

    return p0
.end method

.method private static isCollapsed(Landroid/view/View;)Z
    .locals 1

    .line 102
    invoke-virtual {p0}, Landroid/view/View;->getTag()Ljava/lang/Object;

    move-result-object p0

    .line 103
    instance-of v0, p0, Ljava/lang/Boolean;

    if-eqz v0, :cond_0

    check-cast p0, Ljava/lang/Boolean;

    invoke-virtual {p0}, Ljava/lang/Boolean;->booleanValue()Z

    move-result p0

    if-eqz p0, :cond_0

    const/4 p0, 0x1

    goto :goto_0

    :cond_0
    const/4 p0, 0x0

    :goto_0
    return p0
.end method

.method static synthetic lambda$attach$0(Landroid/app/Activity;Ljava/lang/String;Lcom/gogi/GogiOverlay$OverlayConfig;)V
    .locals 1

    .line 39
    sget-object v0, Lcom/gogi/GogiOverlay;->overlay:Landroid/view/View;

    if-eqz v0, :cond_0

    .line 40
    return-void

    .line 42
    :cond_0
    invoke-static {p0, p1, p2}, Lcom/gogi/GogiOverlay;->createPanel(Landroid/app/Activity;Ljava/lang/String;Lcom/gogi/GogiOverlay$OverlayConfig;)Landroid/view/View;

    move-result-object p1

    sput-object p1, Lcom/gogi/GogiOverlay;->overlay:Landroid/view/View;

    .line 43
    sget-object p1, Lcom/gogi/GogiOverlay;->overlay:Landroid/view/View;

    sget-object p2, Lcom/gogi/GogiOverlay;->overlay:Landroid/view/View;

    invoke-virtual {p2}, Landroid/view/View;->getLayoutParams()Landroid/view/ViewGroup$LayoutParams;

    move-result-object p2

    invoke-virtual {p0, p1, p2}, Landroid/app/Activity;->addContentView(Landroid/view/View;Landroid/view/ViewGroup$LayoutParams;)V

    .line 44
    return-void
.end method

.method static synthetic lambda$createPanel$0(Landroid/webkit/WebView;Ljava/lang/String;)V
    .locals 0

    .line 65
    invoke-virtual {p0, p1}, Landroid/webkit/WebView;->loadUrl(Ljava/lang/String;)V

    return-void
.end method

.method private static setCollapsed(Landroid/view/View;Landroid/view/View;Lcom/gogi/GogiOverlay$OverlayConfig;Z)V
    .locals 2

    .line 93
    const/16 v0, 0x8

    if-eqz p3, :cond_0

    move v1, v0

    goto :goto_0

    :cond_0
    const/4 v1, 0x0

    :goto_0
    invoke-virtual {p1, v1}, Landroid/view/View;->setVisibility(I)V

    .line 94
    invoke-static {p3}, Ljava/lang/Boolean;->valueOf(Z)Ljava/lang/Boolean;

    move-result-object p1

    invoke-virtual {p0, p1}, Landroid/view/View;->setTag(Ljava/lang/Object;)V

    .line 95
    invoke-virtual {p0}, Landroid/view/View;->getLayoutParams()Landroid/view/ViewGroup$LayoutParams;

    move-result-object p1

    check-cast p1, Landroid/widget/FrameLayout$LayoutParams;

    .line 96
    if-eqz p3, :cond_1

    iget v1, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    goto :goto_1

    :cond_1
    iget v1, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->width:I

    :goto_1
    iput v1, p1, Landroid/widget/FrameLayout$LayoutParams;->width:I

    .line 97
    if-eqz p3, :cond_2

    iget p2, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    goto :goto_2

    :cond_2
    iget p3, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    add-int/2addr p3, v0

    iget p2, p2, Lcom/gogi/GogiOverlay$OverlayConfig;->height:I

    add-int/2addr p2, p3

    :goto_2
    iput p2, p1, Landroid/widget/FrameLayout$LayoutParams;->height:I

    .line 98
    invoke-virtual {p0, p1}, Landroid/view/View;->setLayoutParams(Landroid/view/ViewGroup$LayoutParams;)V

    .line 99
    return-void
.end method
