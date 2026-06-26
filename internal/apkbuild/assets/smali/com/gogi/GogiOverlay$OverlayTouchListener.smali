.class final Lcom/gogi/GogiOverlay$OverlayTouchListener;
.super Ljava/lang/Object;
.source "GogiOverlay.java"

# interfaces
.implements Landroid/view/View$OnTouchListener;


# annotations
.annotation system Ldalvik/annotation/EnclosingClass;
    value = Lcom/gogi/GogiOverlay;
.end annotation

.annotation system Ldalvik/annotation/InnerClass;
    accessFlags = 0x1a
    name = "OverlayTouchListener"
.end annotation


# instance fields
.field private final body:Landroid/view/View;

.field private final config:Lcom/gogi/GogiOverlay$OverlayConfig;

.field private downRawX:F

.field private downRawY:F

.field private dragging:Z

.field private startLeft:I

.field private startTop:I

.field private final target:Landroid/view/View;

.field private final touchSlop:I


# direct methods
.method constructor <init>(Landroid/view/View;Landroid/view/View;Lcom/gogi/GogiOverlay$OverlayConfig;)V
    .locals 0

    .line 117
    invoke-direct {p0}, Ljava/lang/Object;-><init>()V

    .line 118
    iput-object p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->target:Landroid/view/View;

    .line 119
    iput-object p2, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->body:Landroid/view/View;

    .line 120
    iput-object p3, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->config:Lcom/gogi/GogiOverlay$OverlayConfig;

    .line 121
    invoke-virtual {p1}, Landroid/view/View;->getContext()Landroid/content/Context;

    move-result-object p1

    invoke-static {p1}, Landroid/view/ViewConfiguration;->get(Landroid/content/Context;)Landroid/view/ViewConfiguration;

    move-result-object p1

    invoke-virtual {p1}, Landroid/view/ViewConfiguration;->getScaledTouchSlop()I

    move-result p1

    iput p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->touchSlop:I

    .line 122
    return-void
.end method

.method private clamp(Landroid/widget/FrameLayout$LayoutParams;)V
    .locals 4

    .line 164
    iget-object v0, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->target:Landroid/view/View;

    invoke-virtual {v0}, Landroid/view/View;->getParent()Landroid/view/ViewParent;

    move-result-object v0

    check-cast v0, Landroid/view/View;

    .line 165
    if-eqz v0, :cond_1

    invoke-virtual {v0}, Landroid/view/View;->getWidth()I

    move-result v1

    if-eqz v1, :cond_1

    invoke-virtual {v0}, Landroid/view/View;->getHeight()I

    move-result v1

    if-nez v1, :cond_0

    goto :goto_0

    .line 168
    :cond_0
    invoke-virtual {v0}, Landroid/view/View;->getWidth()I

    move-result v1

    iget-object v2, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->config:Lcom/gogi/GogiOverlay$OverlayConfig;

    iget v2, v2, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    sub-int/2addr v1, v2

    const/4 v2, 0x0

    invoke-static {v2, v1}, Ljava/lang/Math;->max(II)I

    move-result v1

    .line 169
    invoke-virtual {v0}, Landroid/view/View;->getHeight()I

    move-result v0

    iget-object v3, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->config:Lcom/gogi/GogiOverlay$OverlayConfig;

    iget v3, v3, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    sub-int/2addr v0, v3

    invoke-static {v2, v0}, Ljava/lang/Math;->max(II)I

    move-result v0

    .line 170
    iget v3, p1, Landroid/widget/FrameLayout$LayoutParams;->leftMargin:I

    invoke-static {v3, v1}, Ljava/lang/Math;->min(II)I

    move-result v1

    invoke-static {v2, v1}, Ljava/lang/Math;->max(II)I

    move-result v1

    iput v1, p1, Landroid/widget/FrameLayout$LayoutParams;->leftMargin:I

    .line 171
    iget v1, p1, Landroid/widget/FrameLayout$LayoutParams;->topMargin:I

    invoke-static {v1, v0}, Ljava/lang/Math;->min(II)I

    move-result v0

    invoke-static {v2, v0}, Ljava/lang/Math;->max(II)I

    move-result v0

    iput v0, p1, Landroid/widget/FrameLayout$LayoutParams;->topMargin:I

    .line 172
    return-void

    .line 166
    :cond_1
    :goto_0
    return-void
.end method


# virtual methods
.method public onTouch(Landroid/view/View;Landroid/view/MotionEvent;)Z
    .locals 6

    .line 126
    iget-object v0, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->target:Landroid/view/View;

    invoke-virtual {v0}, Landroid/view/View;->getLayoutParams()Landroid/view/ViewGroup$LayoutParams;

    move-result-object v0

    check-cast v0, Landroid/widget/FrameLayout$LayoutParams;

    .line 127
    invoke-virtual {p2}, Landroid/view/MotionEvent;->getActionMasked()I

    move-result v1

    const/4 v2, 0x0

    const/4 v3, 0x1

    packed-switch v1, :pswitch_data_0

    .line 159
    return v2

    .line 136
    :pswitch_0
    iget-object p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->config:Lcom/gogi/GogiOverlay$OverlayConfig;

    iget-boolean p1, p1, Lcom/gogi/GogiOverlay$OverlayConfig;->draggable:Z

    if-nez p1, :cond_0

    .line 137
    return v3

    .line 139
    :cond_0
    invoke-virtual {p2}, Landroid/view/MotionEvent;->getRawX()F

    move-result p1

    iget v1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->downRawX:F

    sub-float/2addr p1, v1

    invoke-static {p1}, Ljava/lang/Math;->round(F)I

    move-result p1

    .line 140
    invoke-virtual {p2}, Landroid/view/MotionEvent;->getRawY()F

    move-result p2

    iget v1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->downRawY:F

    sub-float/2addr p2, v1

    invoke-static {p2}, Ljava/lang/Math;->round(F)I

    move-result p2

    .line 141
    iget-boolean v1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->dragging:Z

    if-nez v1, :cond_1

    int-to-double v1, p1

    int-to-double v4, p2

    invoke-static {v1, v2, v4, v5}, Ljava/lang/Math;->hypot(DD)D

    move-result-wide v1

    iget v4, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->touchSlop:I

    int-to-double v4, v4

    cmpg-double v1, v1, v4

    if-gez v1, :cond_1

    .line 142
    return v3

    .line 144
    :cond_1
    iput-boolean v3, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->dragging:Z

    .line 145
    iget v1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->startLeft:I

    add-int/2addr v1, p1

    iput v1, v0, Landroid/widget/FrameLayout$LayoutParams;->leftMargin:I

    .line 146
    iget p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->startTop:I

    add-int/2addr p1, p2

    iput p1, v0, Landroid/widget/FrameLayout$LayoutParams;->topMargin:I

    .line 147
    invoke-direct {p0, v0}, Lcom/gogi/GogiOverlay$OverlayTouchListener;->clamp(Landroid/widget/FrameLayout$LayoutParams;)V

    .line 148
    iget-object p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->target:Landroid/view/View;

    invoke-virtual {p1, v0}, Landroid/view/View;->setLayoutParams(Landroid/view/ViewGroup$LayoutParams;)V

    .line 149
    return v3

    .line 152
    :pswitch_1
    iget-boolean p2, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->dragging:Z

    if-nez p2, :cond_2

    .line 153
    iget-object p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->target:Landroid/view/View;

    iget-object p2, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->body:Landroid/view/View;

    iget-object v0, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->config:Lcom/gogi/GogiOverlay$OverlayConfig;

    iget-object v1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->target:Landroid/view/View;

    invoke-static {v1}, Lcom/gogi/GogiOverlay;->-$$Nest$smisCollapsed(Landroid/view/View;)Z

    move-result v1

    xor-int/2addr v1, v3

    invoke-static {p1, p2, v0, v1}, Lcom/gogi/GogiOverlay;->-$$Nest$smsetCollapsed(Landroid/view/View;Landroid/view/View;Lcom/gogi/GogiOverlay$OverlayConfig;Z)V

    goto :goto_0

    .line 155
    :cond_2
    invoke-virtual {p1}, Landroid/view/View;->performClick()Z

    .line 157
    :goto_0
    return v3

    .line 129
    :pswitch_2
    iget p1, v0, Landroid/widget/FrameLayout$LayoutParams;->leftMargin:I

    iput p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->startLeft:I

    .line 130
    iget p1, v0, Landroid/widget/FrameLayout$LayoutParams;->topMargin:I

    iput p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->startTop:I

    .line 131
    invoke-virtual {p2}, Landroid/view/MotionEvent;->getRawX()F

    move-result p1

    iput p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->downRawX:F

    .line 132
    invoke-virtual {p2}, Landroid/view/MotionEvent;->getRawY()F

    move-result p1

    iput p1, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->downRawY:F

    .line 133
    iput-boolean v2, p0, Lcom/gogi/GogiOverlay$OverlayTouchListener;->dragging:Z

    .line 134
    return v3

    :pswitch_data_0
    .packed-switch 0x0
        :pswitch_2
        :pswitch_1
        :pswitch_0
        :pswitch_1
    .end packed-switch
.end method
