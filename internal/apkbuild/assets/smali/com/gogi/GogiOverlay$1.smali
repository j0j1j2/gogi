.class public final synthetic Lcom/gogi/GogiOverlay$1;
.super Ljava/lang/Object;
.source "D8$$SyntheticClass"

# interfaces
.implements Ljava/lang/Runnable;


# annotations
.annotation runtime Lcom/android/tools/r8/annotations/LambdaMethod;
    holder = "Lcom/gogi/GogiOverlay;"
    method = "lambda$createPanel$0"
    proto = "(Landroid/webkit/WebView;Ljava/lang/String;)V"
.end annotation


# instance fields
.field public final synthetic f$0:Landroid/webkit/WebView;

.field public final synthetic f$1:Ljava/lang/String;


# direct methods
.method public synthetic constructor <init>(Landroid/webkit/WebView;Ljava/lang/String;)V
    .locals 0

    .line 0
    invoke-direct {p0}, Ljava/lang/Object;-><init>()V

    iput-object p1, p0, Lcom/gogi/GogiOverlay$1;->f$0:Landroid/webkit/WebView;

    iput-object p2, p0, Lcom/gogi/GogiOverlay$1;->f$1:Ljava/lang/String;

    return-void
.end method


# virtual methods
.method public final run()V
    .locals 2

    .line 0
    iget-object v0, p0, Lcom/gogi/GogiOverlay$1;->f$0:Landroid/webkit/WebView;

    iget-object v1, p0, Lcom/gogi/GogiOverlay$1;->f$1:Ljava/lang/String;

    invoke-static {v0, v1}, Lcom/gogi/GogiOverlay;->lambda$createPanel$0(Landroid/webkit/WebView;Ljava/lang/String;)V

    return-void
.end method
