import reflex as rx

from ovaimporter.templates import template
from .ovauploading import OvaUploadingState
import ovaimporter.commands as cmd

import time

class OvaProcessingState(rx.State):

    progress: int = 50
    
    Stages = {"Extracting files from OVA ...":cmd.untar, 
              "Parsing OVF to collect HW requirements ... ":cmd.parseOvf, 
              "Converting image":cmd.convert, 
              "Uploading image to PVC":cmd.upload,
              "Creatiing boot volume from PVC with uploaded image":cmd.createBootVolume,
              "cleanup":cmd.cleanup
              }

    stage = 0

    finished: bool = False

    messages = []

    async def start_process(self):
        workdir = rx.get_upload_dir()
        for file in workdir.iterdir():
            if file.is_file() and file.name.find('.ova') >= 0:
                print("Processing: {}".format(str(file)))
                for msg, command in self.Stages.items():
                    self.messages.append(msg +":" + file.name)
                    yield
                    r = command(str(workdir), file.name)
                    if isinstance(r, dict):
                        for k,v in r.items():
                            self.messages.append(f"\t{k} : {v}")
                    elif isinstance(r, bool):
                        self.finished = r
                    # yield

    def backToMainPage(self):
        self.messages = []
        self.stage = 0
        return rx.redirect("/")

    @rx.var
    def workInProgress(self) -> bool:
        return len(self.messages) > 0

    @rx.var
    def working(self) -> bool:
        return not self.finished
        # return len(self.messages) < len(self.Stages)
    
def message(message):
    """Display for an individual message in the feed."""
    return rx.grid(
        rx.vstack(
            # rx.text("@" + message.author, font_weight="bold"),
            # rx.text(message.content, width="100%"),
            rx.text(message, width="100%"),
            align_items="left",
        ),
        # grid_template_columns="1fr 5fr",
        padding="1.5rem",
        spacing="1",
        # border_bottom="1px solid #ededed",
    )

@rx.page(route="/processova", on_load=OvaProcessingState.start_process)
@template
def ovaProcessingPage():
    return rx.vstack(
            rx.box(
                rx.heading("Processing OVA ..."), #,style=style.topic_style),
                rx.foreach(
                    OvaProcessingState.messages,
                    message,
                ),
                align="left",
            ),
            # rx.progress(value=OvaProcessingState.progress, max=100),
            rx.button("Next OVA", on_click=OvaProcessingState.backToMainPage,disabled=OvaProcessingState.working),
        ),

