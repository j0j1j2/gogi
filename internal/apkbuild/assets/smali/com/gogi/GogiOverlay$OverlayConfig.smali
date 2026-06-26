.class final Lcom/gogi/GogiOverlay$OverlayConfig;
.super Ljava/lang/Object;
.source "GogiOverlay.java"


# annotations
.annotation system Ldalvik/annotation/EnclosingClass;
    value = Lcom/gogi/GogiOverlay;
.end annotation

.annotation system Ldalvik/annotation/InnerClass;
    accessFlags = 0x1a
    name = "OverlayConfig"
.end annotation


# instance fields
.field final draggable:Z

.field final handleSize:I

.field final height:I

.field final width:I


# direct methods
.method constructor <init>(IIIZ)V
    .locals 0

    .line 181
    invoke-direct {p0}, Ljava/lang/Object;-><init>()V

    .line 182
    iput p1, p0, Lcom/gogi/GogiOverlay$OverlayConfig;->width:I

    .line 183
    iput p2, p0, Lcom/gogi/GogiOverlay$OverlayConfig;->height:I

    .line 184
    iput p3, p0, Lcom/gogi/GogiOverlay$OverlayConfig;->handleSize:I

    .line 185
    iput-boolean p4, p0, Lcom/gogi/GogiOverlay$OverlayConfig;->draggable:Z

    .line 186
    return-void
.end method

.method static parse(Landroid/content/Context;Ljava/lang/String;)Lcom/gogi/GogiOverlay$OverlayConfig;
    .locals 9

    .line 190
    const/4 v0, 0x1

    const/16 v1, 0x38

    const/16 v2, 0x1a4

    const/16 v3, 0x140

    :try_start_0
    new-instance v4, Lorg/json/JSONObject;

    if-nez p1, :cond_0

    const-string p1, "{}"

    :cond_0
    invoke-direct {v4, p1}, Lorg/json/JSONObject;-><init>(Ljava/lang/String;)V

    .line 191
    new-instance p1, Lcom/gogi/GogiOverlay$OverlayConfig;

    const-string v5, "width"

    .line 192
    invoke-virtual {v4, v5, v3}, Lorg/json/JSONObject;->optInt(Ljava/lang/String;I)I

    move-result v5

    invoke-static {p0, v5}, Lcom/gogi/GogiOverlay;->-$$Nest$smdp(Landroid/content/Context;I)I

    move-result v5

    const-string v6, "height"

    .line 193
    invoke-virtual {v4, v6, v2}, Lorg/json/JSONObject;->optInt(Ljava/lang/String;I)I

    move-result v6

    invoke-static {p0, v6}, Lcom/gogi/GogiOverlay;->-$$Nest$smdp(Landroid/content/Context;I)I

    move-result v6

    const-string v7, "collapsed_size"

    .line 194
    invoke-virtual {v4, v7, v1}, Lorg/json/JSONObject;->optInt(Ljava/lang/String;I)I

    move-result v7

    invoke-static {p0, v7}, Lcom/gogi/GogiOverlay;->-$$Nest$smdp(Landroid/content/Context;I)I

    move-result v7

    const-string v8, "draggable"

    .line 195
    invoke-virtual {v4, v8, v0}, Lorg/json/JSONObject;->optBoolean(Ljava/lang/String;Z)Z

    move-result v4

    invoke-direct {p1, v5, v6, v7, v4}, Lcom/gogi/GogiOverlay$OverlayConfig;-><init>(IIIZ)V
    :try_end_0
    .catch Ljava/lang/Exception; {:try_start_0 .. :try_end_0} :catch_0

    .line 191
    return-object p1

    .line 197
    :catch_0
    move-exception p1

    .line 198
    new-instance p1, Lcom/gogi/GogiOverlay$OverlayConfig;

    .line 199
    invoke-static {p0, v3}, Lcom/gogi/GogiOverlay;->-$$Nest$smdp(Landroid/content/Context;I)I

    move-result v3

    .line 200
    invoke-static {p0, v2}, Lcom/gogi/GogiOverlay;->-$$Nest$smdp(Landroid/content/Context;I)I

    move-result v2

    .line 201
    invoke-static {p0, v1}, Lcom/gogi/GogiOverlay;->-$$Nest$smdp(Landroid/content/Context;I)I

    move-result p0

    invoke-direct {p1, v3, v2, p0, v0}, Lcom/gogi/GogiOverlay$OverlayConfig;-><init>(IIIZ)V

    .line 198
    return-object p1
.end method
