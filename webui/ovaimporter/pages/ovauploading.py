import os
import logging
import reflex as rx

from ovaimporter.templates import template
from ovaimporter.pages.collectOcpData import OcpData

import ovaimporter.commands as cmd


class RadioGroupState(rx.State):
    item: str = "Local File"
    uselocalfile: bool = False
    url: str = ''

    def set_item(self, item: str):
        self.item = item
        self.uselocalfile = item == "Local File"

class OvaUploadingState(rx.State):
    progress:int = 0
    total_bytes: int = 0
    uploading: bool = False
    finished: bool = False
   
    async def handle_upload(
        self, files: list[rx.UploadFile]
    ):
        for file in files:
            upload_data = await file.read()

            # Save the file.
            outfile = rx.get_upload_dir() / file.filename
            with outfile.open("wb") as file_object:
                file_object.write(upload_data)

            self.total_bytes += len(upload_data)

    def handle_download(self,form_data: dict):
        cmd.download(form_data["ovaurl"])
        self.finished = True

    def handle_upload_progress(self, progress: dict):
        self.uploading = True
        self.progress = round(progress["progress"] * 100)
        if self.progress >= 100:
            self.uploading = False
            self.finished = True
            # self.progress = 0

    # def cancel_upload(self):
    #     self.uploading = False
    #     self.finished = False
    #     return rx.cancel_upload("upload3")

    def continueWithOvaProcessing(self):

        # Reset all fields, and then proceed to next page
        self.finished = False
        self.uploading = False
        self.total_bytes = 0
        self.progress = 0
        OvaUploadingState.progress = 0
        return rx.redirect("/processova")

    def cancel(self):

        # Reset all fields, and then proceed to next page
        self.finished = False
        self.uploading = False
        self.total_bytes = 0
        self.progress = 0
        OvaUploadingState.progress = 0
    
    @rx.var
    def isNotReadyToUpload(self) -> bool:
        return self.progress> 0

    @rx.var
    def isNotFinished(self) -> bool:
        return not self.finished

@rx.page(route="/", on_load=OcpData.isCollected())
@template
def ovaUploadingPage():
    return rx.vstack(
        #---

        rx.vstack(
            rx.heading("OVA source"),
            rx.radio(
                ["Local File", "URL"],
                on_change=RadioGroupState.set_item,
                direction="row",
                default_value="URL"
            ),
        ),
        #---
        rx.cond(RadioGroupState.uselocalfile,
            rx.card(
                rx.vstack(
                    rx.upload(
                        rx.button("Select File"),
                        # rx.text(
                        #     "Drag and drop files here or click to select files"
                        # ),
                        id="upload3",
                        border="1px dotted rgb(107,99,246)",
                        padding="5em",
                        multiple=False,
                        max_files=1,
                    ),
                    rx.vstack(
                        rx.foreach(
                            rx.selected_files("upload3"), rx.text
                        )
                    ),
                    rx.progress(value=OvaUploadingState.progress, max=100),
                    rx.hstack(
                        rx.button(
                            "Upload",
                            disabled=OvaUploadingState.isNotReadyToUpload,
                            on_click=OvaUploadingState.handle_upload(
                                rx.upload_files(
                                    upload_id="upload3",
                                    on_upload_progress=OvaUploadingState.handle_upload_progress,
                                ),
                            ),
                        ),
                        # rx.button("Clear", on_click=rx.clear_selected_files("upload3"), disabled=OvaUploadingState.isNotFinished),
                        # rx.button(
                        #     "Cancel",
                        #     on_click=OvaUploadingState.cancel_upload,
                        # ),
                    ),
                    rx.text(
                        "Total bytes uploaded: ",
                        OvaUploadingState.total_bytes,
                    ),
                    align="center",
                )
            ),
            # rx.cond FALSE -------------------------------------
            rx.card(
                rx.vstack(
                    # rx.heading("OVA URL"),
                    rx.form.root(
                        rx.hstack(
                            rx.input(
                                name="ovaurl",
                                placeholder="Enter URL...",
                                type="text",
                                required=True,
                            ),
                            rx.button("Download", type="submit"),
                            width="100%",
                        ),
                        on_submit=OvaUploadingState.handle_download,
                        reset_on_submit=True,
                    ),
                ),
                align_items="left",
                width="100%",
            ),
        ),
        rx.hstack(
            rx.button("Proceed", on_click=OvaUploadingState.continueWithOvaProcessing, disabled=OvaUploadingState.isNotFinished),
            rx.button("Cancel", on_click=OvaUploadingState.cancel, disabled=OvaUploadingState.isNotFinished),
        ),
    ),
